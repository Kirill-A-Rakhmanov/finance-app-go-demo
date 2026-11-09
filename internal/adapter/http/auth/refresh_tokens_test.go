package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpauth "finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newRefreshTokensTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/refresh", h.RefreshTokens)
	return r
}

func TestHandler_RefreshTokens(t *testing.T) {
	tests := []struct {
		name          string
		refreshCookie *string
		setupMocks    func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService)
		wantStatus    int
		wantCookies   bool
	}{
		{
			name:          "successful refresh sets new cookies",
			refreshCookie: strPtr("valid-refresh-token"),
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenData := &auth.TokenData{SessionID: "old-session"}
				tokenService.EXPECT().VerifyToken("valid-refresh-token", auth.RefreshTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "old-session").Return(&auth.Session{ID: "old-session"}, nil)
				sessionStorage.EXPECT().CreateSession(mock.Anything, mock.Anything).Return(&auth.Session{ID: "new-session"}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(&auth.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}, nil)
				sessionStorage.EXPECT().ShortenSession(mock.Anything, mock.Anything, auth.RefreshGracePeriod).Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantCookies: true,
		},
		{
			name:          "missing refresh token cookie returns 401",
			refreshCookie: nil,
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:          "service error clears cookies and returns 401",
			refreshCookie: strPtr("garbage-token"),
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("garbage-token", auth.RefreshTokenType).Return(nil, errors.New("token expired"))
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionStorage := authmock.NewMockSessionStorage(t)
			tokenService := authmock.NewMockTokenService(t)
			tt.setupMocks(sessionStorage, tokenService)

			authService := auth.NewService(nil, nil, nil, tokenService, nil, nil, sessionStorage, nil, nil, nil, nil, nil)
			router := newRefreshTokensTestRouter(authService)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tt.refreshCookie != nil {
				req.AddCookie(&http.Cookie{Name: cookie.RefreshTokenName, Value: *tt.refreshCookie})
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantCookies {
				var cookieNames []string
				for _, c := range rec.Result().Cookies() {
					cookieNames = append(cookieNames, c.Name)
				}
				assert.Contains(t, cookieNames, cookie.AccessTokenName)
				assert.Contains(t, cookieNames, cookie.RefreshTokenName)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
