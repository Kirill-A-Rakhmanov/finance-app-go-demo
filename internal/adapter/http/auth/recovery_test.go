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

func newRecoveryTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/recovery", h.Recovery)
	return r
}

func TestHandler_Recovery(t *testing.T) {
	email, err := valueObject.NewEmail("jane@example.com")
	require.NoError(t, err)
	user := &model.User{ID: uuid.New(), FirstName: "Jane", Email: email, Password: "hashed-password"}

	tests := []struct {
		name       string
		body       string
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			recoveryStorage *authmock.MockRecoveryStorage,
			recoveryNotifier *authmock.MockRecoveryNotifier,
		)
		wantStatus int
	}{
		{
			name: "successful recovery request",
			body: `{"email":"jane@example.com"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				recoveryNotifier *authmock.MockRecoveryNotifier,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "jane@example.com").Return(user, nil)
				recoveryStorage.EXPECT().CreateRecovery(mock.Anything, user.ID).Return(&auth.Recovery{Code: "recovery-code", UserID: user.ID}, nil)
				recoveryNotifier.EXPECT().SendPasswordRecovery(mock.Anything, "jane@example.com", "recovery-code").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "malformed request body returns 400",
			body: `{"email": }`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error returns 422",
			body: `{"email":""}`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "unknown email gets the same 200 as a known one",
			body: `{"email":"unknown@example.com"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				_ *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "unknown@example.com").Return(nil, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			recoveryStorage := authmock.NewMockRecoveryStorage(t)
			recoveryNotifier := authmock.NewMockRecoveryNotifier(t)
			recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, mock.Anything, auth.RecoveryCooldown).Return(true, nil).Maybe()
			tt.setupMocks(userRepo, recoveryStorage, recoveryNotifier)

			authService := auth.NewService(userRepo, nil, nil, nil, nil, nil, nil, recoveryStorage, recoveryNotifier, nil, nil, nil)
			router := newRecoveryTestRouter(authService)

			req := httptest.NewRequest(http.MethodPost, "/auth/recovery", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
