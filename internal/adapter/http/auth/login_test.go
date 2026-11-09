package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpauth "finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newLoginTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/login", h.Login)
	return r
}

func TestHandler_Login(t *testing.T) {
	email, err := valueObject.NewEmail("jane@example.com")
	require.NoError(t, err)
	user := &model.User{ID: uuid.New(), FirstName: "Jane", Email: email, Password: "hashed-password"}

	tests := []struct {
		name       string
		body       string
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			passwordService *authmock.MockPasswordService,
			sessionStorage *authmock.MockSessionStorage,
			tokenService *authmock.MockTokenService,
		)
		retryAfter time.Duration
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "successful login sets cookies and returns user",
			body: `{"email":"jane@example.com","password":"correct-password"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "jane@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("correct-password", "hashed-password").Return(true)
				sessionStorage.EXPECT().CreateSession(mock.Anything, user.ID).Return(&auth.Session{ID: "session-1", UserID: user.ID}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(&auth.TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				userBody := body["user"].(map[string]any)
				assert.Equal(t, "jane@example.com", userBody["email"])

				var cookieNames []string
				for _, c := range rec.Result().Cookies() {
					cookieNames = append(cookieNames, c.Name)
				}
				assert.Contains(t, cookieNames, "access_token")
				assert.Contains(t, cookieNames, "refresh_token")
			},
		},
		{
			name: "malformed request body returns 400",
			body: `{"email": "jane@example.com", "password": }`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "business error maps through response.FromError",
			body: `{"email":"jane@example.com","password":"wrong-password"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "jane@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("wrong-password", "hashed-password").Return(false)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "locked account maps to 429 with Retry-After",
			body:       `{"email":"jane@example.com","password":"whatever"}`,
			retryAfter: 90 * time.Second,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
			},
			wantStatus: http.StatusTooManyRequests,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Equal(t, "90", rec.Header().Get("Retry-After"))

				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.NotEmpty(t, body["error"])
			},
		},
		{
			name: "validation error returns 422 with field detail",
			body: `{"email":"","password":""}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				fields := body["fields"].(map[string]any)
				assert.Contains(t, fields, "email")
				assert.Contains(t, fields, "password")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			passwordService := authmock.NewMockPasswordService(t)
			sessionStorage := authmock.NewMockSessionStorage(t)
			tokenService := authmock.NewMockTokenService(t)
			limiter := authmock.NewMockFailureLimiter(t)
			limiter.EXPECT().RetryAfter(mock.Anything, mock.Anything, auth.MaxFailedAttempts).Return(tt.retryAfter, nil).Maybe()
			limiter.EXPECT().RecordFailure(mock.Anything, mock.Anything, auth.FailedAttemptsWindow).Return(nil).Maybe()
			limiter.EXPECT().Reset(mock.Anything, mock.Anything).Return(nil).Maybe()
			tt.setupMocks(userRepo, passwordService, sessionStorage, tokenService)

			authService := auth.NewService(userRepo, nil, nil, tokenService, passwordService, nil, sessionStorage, nil, nil, nil, nil, limiter)
			router := newLoginTestRouter(authService)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
