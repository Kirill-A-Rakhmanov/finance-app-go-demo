package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

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

func newResetPasswordTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/reset-password", h.ResetPassword)
	return r
}

func TestHandler_ResetPassword(t *testing.T) {
	email, err := valueObject.NewEmail("jane@example.com")
	require.NoError(t, err)
	user := &model.User{ID: uuid.New(), FirstName: "Jane", Email: email, Password: "old-hashed-password"}

	tests := []struct {
		name       string
		body       string
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			passwordService *authmock.MockPasswordService,
			recoveryStorage *authmock.MockRecoveryStorage,
		)
		wantStatus int
	}{
		{
			name: "successful reset",
			body: `{"code":"recovery-code","newPassword":"NewPassword123"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: user.ID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, user.ID).Return(user, nil)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(user, nil)
				recoveryStorage.EXPECT().DeleteRecovery(mock.Anything, "recovery-code").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "malformed request body returns 400",
			body: `{"code": }`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRecoveryStorage,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error returns 422",
			body: `{"code":"","newPassword":"short"}`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRecoveryStorage,
			) {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid recovery code maps through response.FromError",
			body: `{"code":"unknown-code","newPassword":"NewPassword123"}`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "unknown-code").Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			passwordService := authmock.NewMockPasswordService(t)
			recoveryStorage := authmock.NewMockRecoveryStorage(t)
			sessionStorage := authmock.NewMockSessionStorage(t)
			sessionStorage.EXPECT().DeleteSessionsByUserID(mock.Anything, user.ID, "").Return(nil).Maybe()
			limiter := authmock.NewMockFailureLimiter(t)
			limiter.EXPECT().Reset(mock.Anything, mock.Anything).Return(nil).Maybe()
			tt.setupMocks(userRepo, passwordService, recoveryStorage)

			authService := auth.NewService(userRepo, nil, nil, nil, passwordService, nil, sessionStorage, recoveryStorage, nil, nil, nil, limiter)
			router := newResetPasswordTestRouter(authService)

			req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
