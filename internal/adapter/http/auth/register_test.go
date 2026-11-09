package auth_test

import (
	"bytes"
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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRegisterTestRouter(authService *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := httpauth.NewHandler(authService, cookie.Config{})
	r.POST("/auth/register", h.Register)
	return r
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		registration bool
		setupMocks   func(
			userRepo *repomock.MockUserRepository,
			passwordService *authmock.MockPasswordService,
			storage *authmock.MockRegistrationStorage,
			notifier *authmock.MockRegistrationNotifier,
		)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:         "successful registration asks to check the email and sets no cookies",
			body:         `{"firstName":"Jane","lastName":"Doe","email":"jane@example.com","password":"Password123"}`,
			registration: true,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				storage *authmock.MockRegistrationStorage,
				notifier *authmock.MockRegistrationNotifier,
			) {
				passwordService.EXPECT().Hash("Password123").Return("hashed-password", nil)
				storage.EXPECT().AcquireCooldown(mock.Anything, "jane@example.com", auth.RegistrationCooldown).Return(true, nil)
				userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(false, nil)
				storage.EXPECT().Create(mock.Anything, mock.Anything).Return("confirmation-token", nil)
				notifier.EXPECT().SendRegistrationConfirmation(mock.Anything, mock.Anything).Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "check your email to finish the registration", body["message"])
				assert.Empty(t, rec.Result().Cookies(), "nobody is signed in before the email is confirmed")
				assert.NotContains(t, body, "user")
			},
		},
		{
			name:         "taken email gets the same response",
			body:         `{"firstName":"Jane","lastName":"Doe","email":"jane@example.com","password":"Password123"}`,
			registration: true,
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				storage *authmock.MockRegistrationStorage,
				notifier *authmock.MockRegistrationNotifier,
			) {
				passwordService.EXPECT().Hash("Password123").Return("hashed-password", nil)
				storage.EXPECT().AcquireCooldown(mock.Anything, "jane@example.com", auth.RegistrationCooldown).Return(true, nil)
				userRepo.EXPECT().ExistsByEmail(mock.Anything, "jane@example.com").Return(true, nil)
				notifier.EXPECT().SendRegistrationAttempt(mock.Anything, "jane@example.com").Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "check your email to finish the registration", body["message"])
			},
		},
		{
			name:         "registration closed returns 403",
			body:         `{"firstName":"Jane","lastName":"Doe","email":"jane@example.com","password":"Password123"}`,
			registration: false,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRegistrationStorage,
				_ *authmock.MockRegistrationNotifier,
			) {
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:         "malformed request body returns 400",
			body:         `{"firstName": }`,
			registration: true,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRegistrationStorage,
				_ *authmock.MockRegistrationNotifier,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:         "validation error returns 422 with field details",
			body:         `{"firstName":"","lastName":"","email":"","password":"short"}`,
			registration: true,
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRegistrationStorage,
				_ *authmock.MockRegistrationNotifier,
			) {
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				fields := body["fields"].(map[string]any)
				assert.Contains(t, fields, "firstName")
				assert.Contains(t, fields, "lastName")
				assert.Contains(t, fields, "email")
				assert.Contains(t, fields, "password")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			featureFlagRepo := repomock.NewMockFeatureFlagRepository(t)
			passwordService := authmock.NewMockPasswordService(t)
			storage := authmock.NewMockRegistrationStorage(t)
			notifier := authmock.NewMockRegistrationNotifier(t)
			featureFlagRepo.EXPECT().IsEnabled(mock.Anything, model.FeatureFlagRegistration).Return(tt.registration, nil).Maybe()
			tt.setupMocks(userRepo, passwordService, storage, notifier)

			authService := auth.NewService(userRepo, nil, featureFlagRepo, nil, passwordService, nil, nil, nil, nil, storage, notifier, nil)
			router := newRegisterTestRouter(authService)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tt.body))
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
