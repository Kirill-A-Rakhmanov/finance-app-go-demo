package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/adapter/http/middleware"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	"finance-app-go-demo/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_PutsTheSessionAndLogFieldsIntoTheContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	tokenService := authmock.NewMockTokenService(t)
	sessionStorage := authmock.NewMockSessionStorage(t)
	tokenService.EXPECT().VerifyToken("access-token", auth.AccessTokenType).Return(&auth.TokenData{UserID: userID, SessionID: "session-1"}, nil)
	sessionStorage.EXPECT().GetSession(mock.Anything, "session-1").Return(&auth.Session{ID: "session-1", UserID: userID}, nil)

	authService := auth.NewService(nil, nil, nil, tokenService, nil, nil, sessionStorage, nil, nil, nil, nil, nil)

	var buf bytes.Buffer
	log := logger.New(slog.New(logger.NewContextHandler(slog.NewJSONHandler(&buf, nil))))

	var seen request.Session
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.NewAuthMiddleware(authService).Handler())
	r.GET("/me", func(c *gin.Context) {
		seen, _ = request.CurrentSession(c.Request.Context())
		log.InfoContext(c.Request.Context(), "inside the handler")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: cookie.AccessTokenName, Value: "access-token"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, userID, seen.UserID)

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	assert.Equal(t, userID.String(), line["userID"])
	assert.Equal(t, "session-1", line["sessionID"])
	assert.Equal(t, rec.Header().Get(middleware.RequestIDHeader), line["requestID"])
}

func TestAuthMiddleware_RejectsRequestsWithoutAValidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tokenService := authmock.NewMockTokenService(t)
	tokenService.EXPECT().VerifyToken("bad-token", auth.AccessTokenType).Return(nil, auth.ErrInvalidSession)
	authService := auth.NewService(nil, nil, nil, tokenService, nil, nil, nil, nil, nil, nil, nil, nil)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.NewAuthMiddleware(authService).Handler())
	r.GET("/me", func(c *gin.Context) { c.Status(http.StatusOK) })

	t.Run("no cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader), "the id is also on rejected requests")
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.AddCookie(&http.Cookie{Name: cookie.AccessTokenName, Value: "bad-token"})
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
