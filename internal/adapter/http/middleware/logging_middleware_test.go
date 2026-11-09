package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"finance-app-go-demo/internal/adapter/http/middleware"
	"finance-app-go-demo/internal/adapter/http/response"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/pkg/logger"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingLogger struct {
	entries   []string
	errorArgs [][]any
}

func (l *recordingLogger) Info(msg string, _ ...any)  { l.entries = append(l.entries, "info") }
func (l *recordingLogger) Warn(msg string, _ ...any)  { l.entries = append(l.entries, "warn") }
func (l *recordingLogger) Debug(msg string, _ ...any) { l.entries = append(l.entries, "debug") }

func (l *recordingLogger) Error(msg string, args ...any) {
	l.entries = append(l.entries, "error")
	l.errorArgs = append(l.errorArgs, args)
}

func (l *recordingLogger) InfoContext(_ context.Context, msg string, args ...any) {
	l.Info(msg, args...)
}

func (l *recordingLogger) WarnContext(_ context.Context, msg string, args ...any) {
	l.Warn(msg, args...)
}

func (l *recordingLogger) DebugContext(_ context.Context, msg string, args ...any) {
	l.Debug(msg, args...)
}

func (l *recordingLogger) ErrorContext(_ context.Context, msg string, args ...any) {
	l.Error(msg, args...)
}

func TestLoggingMiddleware_QuietPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		path    string
		status  int
		wantLog []string
	}{
		{"successful health check is not logged", "/health", http.StatusOK, nil},
		{"failed health check is logged as an error", "/health", http.StatusServiceUnavailable, []string{"error"}},
		{"client error on a quiet path is logged as a warning", "/health", http.StatusBadRequest, []string{"warn"}},
		{"other paths are always logged", "/api/accounts", http.StatusOK, []string{"info"}},
		{"unknown paths are logged too", "/nope", http.StatusNotFound, []string{"warn"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := &recordingLogger{}

			r := gin.New()
			r.Use(middleware.NewLoggingMiddleware(log).Handler())
			r.GET("/health", func(c *gin.Context) { c.Status(tt.status) })
			r.GET("/api/accounts", func(c *gin.Context) { c.Status(tt.status) })

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			assert.Equal(t, tt.wantLog, log.entries)
		})
	}
}

func TestLoggingMiddleware_RecoveryLogsPanicAndReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	log := &recordingLogger{}
	lm := middleware.NewLoggingMiddleware(log)

	r := gin.New()
	r.Use(lm.Handler(), lm.Recovery())
	r.GET("/boom", func(c *gin.Context) { panic("something broke") })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":"internal server error"}`, rec.Body.String())
	// the panic itself, then the request line with status 500
	assert.Equal(t, []string{"error", "error"}, log.entries)
	require.Len(t, log.errorArgs, 2)
	assert.Contains(t, log.errorArgs[0], "something broke")
	assert.Contains(t, log.errorArgs[0], "/boom")
}

func TestLoggingMiddleware_RecoveryRepanicsAbortHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	lm := middleware.NewLoggingMiddleware(&recordingLogger{})

	r := gin.New()
	r.Use(lm.Recovery())
	r.GET("/abort", func(c *gin.Context) { panic(http.ErrAbortHandler) })

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
	})
}

func TestLoggingMiddleware_EveryLineOfARequestCarriesItsContextFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	log := logger.New(slog.New(logger.NewContextHandler(slog.NewJSONHandler(&buf, nil))))

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.NewLoggingMiddleware(log).Handler())
	r.GET("/ping", func(c *gin.Context) {
		log.InfoContext(c.Request.Context(), "inside the handler")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(middleware.RequestIDHeader, "client-id-12345")
	r.ServeHTTP(httptest.NewRecorder(), req)

	decoder := json.NewDecoder(&buf)
	var messages []string
	for decoder.More() {
		var line map[string]any
		require.NoError(t, decoder.Decode(&line))
		assert.Equal(t, "client-id-12345", line["requestID"], line["msg"])
		messages = append(messages, line["msg"].(string))
	}

	assert.Equal(t, []string{"inside the handler", "http request"}, messages)
}

func TestLoggingMiddleware_LogsTheRealCauseOfAServerErrorButNotTheClientMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	log := logger.New(slog.New(slog.NewJSONHandler(&buf, nil)))

	r := gin.New()
	r.Use(middleware.NewLoggingMiddleware(log).Handler())
	r.GET("/fail", func(c *gin.Context) {
		response.FromError(c.Writer, errors.New("pq: connection reset"))
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fail", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "connection reset")

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "pq: connection reset", line["error"])
	assert.EqualValues(t, http.StatusInternalServerError, line["status"])
	assert.Equal(t, "ERROR", line["level"])
}

func TestLoggingMiddleware_ClientErrorKeepsItsOwnMessageInTheLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	log := logger.New(slog.New(slog.NewJSONHandler(&buf, nil)))

	r := gin.New()
	r.Use(middleware.NewLoggingMiddleware(log).Handler())
	r.GET("/conflict", func(c *gin.Context) {
		response.FromError(c.Writer, apperror.Conflict("user already exists"))
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/conflict", nil))

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, "user already exists", line["error"])
}
