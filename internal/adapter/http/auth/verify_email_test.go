package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpauth "finance-app-go-demo/internal/adapter/http/auth"
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newVerifyEmailTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/verify-email", h.VerifyEmail)
	return r
}

func TestHandler_VerifyEmail(t *testing.T) {
	pending := &auth.PendingRegistration{
		FirstName: "Jane", LastName: "Doe", Email: "jane@example.com", PasswordHash: "hashed-password",
	}

	tests := []struct {
		name       string
		body       string
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			categoryRepo *repomock.MockTransactionCategoryRepository,
			txManager *transactionmock.MockManager,
			storage *authmock.MockRegistrationStorage,
		)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "valid token creates the account without signing in",
			body: `{"token":"confirmation-token"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				categoryRepo *repomock.MockTransactionCategoryRepository,
				txManager *transactionmock.MockManager,
				storage *authmock.MockRegistrationStorage,
			) {
				storage.EXPECT().Get(mock.Anything, "confirmation-token").Return(pending, nil)
				userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				txManager.EXPECT().Wrap(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(&model.User{ID: uuid.New()}, nil)
				categoryRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).Return(nil, nil)
				storage.EXPECT().Delete(mock.Anything, "confirmation-token").Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "registration confirmed, you can log in now", body["message"])
				assert.Empty(t, rec.Result().Cookies())
			},
		},
		{
			name: "unknown or expired token returns 400",
			body: `{"token":"nope"}`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *repomock.MockTransactionCategoryRepository,
				_ *transactionmock.MockManager,
				storage *authmock.MockRegistrationStorage,
			) {
				storage.EXPECT().Get(mock.Anything, "nope").Return(nil, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "already confirmed returns 409",
			body: `{"token":"confirmation-token"}`,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				_ *repomock.MockTransactionCategoryRepository,
				_ *transactionmock.MockManager,
				storage *authmock.MockRegistrationStorage,
			) {
				storage.EXPECT().Get(mock.Anything, "confirmation-token").Return(pending, nil)
				userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(true, nil)
				storage.EXPECT().Delete(mock.Anything, "confirmation-token").Return(nil)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "malformed request body returns 400",
			body: `{"token": }`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *repomock.MockTransactionCategoryRepository,
				_ *transactionmock.MockManager,
				_ *authmock.MockRegistrationStorage,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty token returns 422",
			body: `{"token":""}`,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *repomock.MockTransactionCategoryRepository,
				_ *transactionmock.MockManager,
				_ *authmock.MockRegistrationStorage,
			) {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			categoryRepo := repomock.NewMockTransactionCategoryRepository(t)
			txManager := transactionmock.NewMockManager(t)
			storage := authmock.NewMockRegistrationStorage(t)
			tt.setupMocks(userRepo, categoryRepo, txManager, storage)

			authService := auth.NewService(userRepo, categoryRepo, nil, nil, nil, txManager, nil, nil, nil, storage, nil, nil)
			router := newVerifyEmailTestRouter(authService)

			req := httptest.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewBufferString(tt.body))
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
