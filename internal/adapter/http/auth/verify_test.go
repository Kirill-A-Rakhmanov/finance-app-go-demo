package auth_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpauth "finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newVerifyTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.GET("/auth/verify", h.Verify)
	return r
}

func TestHandler_Verify(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name         string
		accessCookie *string
		setupMocks   func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService)
		wantStatus   int
		wantBody     func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:         "successful verification returns user id",
			accessCookie: strPtr("valid-access-token"),
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenData := &auth.TokenData{UserID: userID, SessionID: "session-1"}
				tokenService.EXPECT().VerifyToken("valid-access-token", auth.AccessTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "session-1").Return(&auth.Session{ID: "session-1", UserID: userID}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, userID.String(), body["userId"])
				assert.Equal(t, userID.String(), rec.Header().Get("X-User-ID"))
			},
		},
		{
			name:         "missing access token cookie returns 401",
			accessCookie: nil,
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:         "service error returns 401",
			accessCookie: strPtr("garbage-token"),
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("garbage-token", auth.AccessTokenType).Return(nil, errors.New("token expired"))
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
			router := newVerifyTestRouter(authService)

			req := httptest.NewRequest(http.MethodGet, "/auth/verify", nil)
			if tt.accessCookie != nil {
				req.AddCookie(&http.Cookie{Name: cookie.AccessTokenName, Value: *tt.accessCookie})
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
