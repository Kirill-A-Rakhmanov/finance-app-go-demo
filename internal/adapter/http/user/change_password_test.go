package user_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"finance-app-go-demo/internal/adapter/http/request"
	httpuser "finance-app-go-demo/internal/adapter/http/user"
	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	"finance-app-go-demo/internal/application/user"
	usermock "finance-app-go-demo/internal/application/user/mock"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newChangePasswordTestRouter(userService *user.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpuser.NewHandler(userService)
	r.PATCH("/user/current/password", h.ChangePassword)
	return r
}

func TestHandler_ChangePassword(t *testing.T) {
	session := &request.Session{UserID: uuid.New(), SessionID: "current-session"}

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService)
		retryAfter time.Duration
		wantStatus int
	}{
		{
			name:    "successful change",
			session: session,
			body:    `{"oldPassword":"old-password","newPassword":"NewPassword123"}`,
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: session.UserID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, session.UserID).Return(existing, nil)
				passwordService.EXPECT().Verify("old-password", "old-hashed-password").Return(true)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(existing, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "no session returns 401",
			session: nil,
			body:    `{"oldPassword":"old-password","newPassword":"NewPassword123"}`,
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"oldPassword": }`,
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "validation error returns 422",
			session: session,
			body:    `{"oldPassword":"","newPassword":"short"}`,
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "locked after repeated wrong passwords maps to 429",
			session:    session,
			body:       `{"oldPassword":"old-password","newPassword":"NewPassword123"}`,
			retryAfter: time.Minute,
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantStatus: http.StatusTooManyRequests,
		},
		{
			name:    "invalid current password maps to 403",
			session: session,
			body:    `{"oldPassword":"wrong-password","newPassword":"NewPassword123"}`,
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: session.UserID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, session.UserID).Return(existing, nil)
				passwordService.EXPECT().Verify("wrong-password", "old-hashed-password").Return(false)
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			passwordService := authmock.NewMockPasswordService(t)
			tt.setupMocks(userRepo, passwordService)

			limiter := authmock.NewMockFailureLimiter(t)
			limiter.EXPECT().RetryAfter(mock.Anything, mock.Anything, auth.MaxFailedAttempts).Return(tt.retryAfter, nil).Maybe()
			limiter.EXPECT().RecordFailure(mock.Anything, mock.Anything, auth.FailedAttemptsWindow).Return(nil).Maybe()
			limiter.EXPECT().Reset(mock.Anything, mock.Anything).Return(nil).Maybe()
			revoker := usermock.NewMockSessionRevoker(t)
			revoker.EXPECT().DeleteSessionsByUserID(mock.Anything, session.UserID, "current-session").Return(nil).Maybe()

			userService := user.NewService(userRepo, nil, passwordService, revoker, limiter)
			router := newChangePasswordTestRouter(userService, tt.session)

			req := httptest.NewRequest(http.MethodPatch, "/user/current/password", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
