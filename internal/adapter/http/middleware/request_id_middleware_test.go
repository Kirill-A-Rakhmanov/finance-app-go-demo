package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finance-app-go-demo/internal/adapter/http/middleware"
	"finance-app-go-demo/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRequestIDRouter(buf *bytes.Buffer) *gin.Engine {
	gin.SetMode(gin.TestMode)

	log := logger.New(slog.New(logger.NewContextHandler(slog.NewJSONHandler(buf, nil))))

	r := gin.New()
	r.Use(middleware.RequestID())
	r.GET("/ping", func(c *gin.Context) {
		log.InfoContext(c.Request.Context(), "handled")
		c.Status(http.StatusOK)
	})
	r.GET("/panic", func(c *gin.Context) { c.AbortWithStatus(http.StatusInternalServerError) })

	return r
}

func serve(r *gin.Engine, path, requestID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if requestID != "" {
		req.Header.Set(middleware.RequestIDHeader, requestID)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func loggedRequestID(t *testing.T, buf *bytes.Buffer) string {
	t.Helper()

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))

	return line["requestID"].(string)
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var buf bytes.Buffer
	rec := serve(newRequestIDRouter(&buf), "/ping", "")

	id := rec.Header().Get(middleware.RequestIDHeader)
	_, err := uuid.Parse(id)
	require.NoError(t, err, "a generated id is a UUID")
	assert.Equal(t, id, loggedRequestID(t, &buf), "the logs carry the id sent back to the client")
}

func TestRequestID_AcceptsAValidIncomingID(t *testing.T) {
	for _, id := range []string{
		"4f9c1a7e2b3d4c5e8f9a0b1c2d3e4f50",
		"web-1.req_42",
		strings.Repeat("a", 64),
		strings.Repeat("a", 8),
	} {
		var buf bytes.Buffer
		rec := serve(newRequestIDRouter(&buf), "/ping", id)

		assert.Equal(t, id, rec.Header().Get(middleware.RequestIDHeader))
		assert.Equal(t, id, loggedRequestID(t, &buf))
	}
}

func TestRequestID_ReplacesAnInvalidIncomingID(t *testing.T) {
	for name, id := range map[string]string{
		"too short":     "short",
		"too long":      strings.Repeat("a", 65),
		"bad character": "abc def ghi jkl",
		"quotes":        `abcdefgh"}{"x":"`,
		"unicode":       "идентификатор-1",
	} {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			rec := serve(newRequestIDRouter(&buf), "/ping", id)

			got := rec.Header().Get(middleware.RequestIDHeader)
			assert.NotEqual(t, id, got)
			_, err := uuid.Parse(got)
			assert.NoError(t, err)
			assert.Equal(t, got, loggedRequestID(t, &buf))
		})
	}
}

func TestRequestID_IsOnEveryKindOfResponse(t *testing.T) {
	var buf bytes.Buffer
	r := newRequestIDRouter(&buf)

	for _, path := range []string{"/ping", "/panic", "/missing"} {
		rec := serve(r, path, "")
		assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader), path)
	}
}

func TestRequestID_DiffersBetweenRequests(t *testing.T) {
	var buf bytes.Buffer
	r := newRequestIDRouter(&buf)

	first := serve(r, "/ping", "").Header().Get(middleware.RequestIDHeader)
	second := serve(r, "/ping", "").Header().Get(middleware.RequestIDHeader)

	assert.NotEqual(t, first, second)
}
