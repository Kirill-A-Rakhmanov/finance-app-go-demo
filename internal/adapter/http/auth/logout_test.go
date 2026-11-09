package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpauth "finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newLogoutTestRouter(authService *auth.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if session != nil {
			ctx := request.WithSession(c.Request.Context(), *session)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	})
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/logout", h.Logout)
	return r
}

func TestHandler_Logout(t *testing.T) {
	session := &request.Session{UserID: uuid.New(), SessionID: "session-1"}

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(sessionStorage *authmock.MockSessionStorage)
		wantStatus int
	}{
		{
			name:    "successful logout clears cookies",
			session: session,
			setupMocks: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSession(mock.Anything, session.SessionID).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "no session in context returns 401",
			session: nil,
			setupMocks: func(sessionStorage *authmock.MockSessionStorage) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "service error returns 500",
			session: session,
			setupMocks: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSession(mock.Anything, session.SessionID).Return(errors.New("redis unavailable"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionStorage := authmock.NewMockSessionStorage(t)
			tt.setupMocks(sessionStorage)

			authService := auth.NewService(nil, nil, nil, nil, nil, nil, sessionStorage, nil, nil, nil, nil, nil)
			router := newLogoutTestRouter(authService, tt.session)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
