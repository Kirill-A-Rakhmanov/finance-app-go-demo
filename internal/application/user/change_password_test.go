package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/user"
	usermock "finance-app-go-demo/internal/application/user/mock"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ChangePassword(t *testing.T) {
	userID := uuid.New()

	key := auth.ChangePasswordFailuresKey(userID)
	notLocked := func(limiter *authmock.MockFailureLimiter) {
		limiter.EXPECT().RetryAfter(mock.Anything, key, auth.MaxFailedAttempts).Return(time.Duration(0), nil)
	}
	passes := func(limiter *authmock.MockFailureLimiter) {
		notLocked(limiter)
		limiter.EXPECT().Reset(mock.Anything, key).Return(nil)
	}

	tests := []struct {
		name         string
		input        user.ChangePasswordInput
		setupMocks   func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService)
		setupRevoker func(revoker *usermock.MockSessionRevoker)
		setupLimiter func(limiter *authmock.MockFailureLimiter)
		wantErr      func(t *testing.T, err error)
		wantResult   func(t *testing.T, result *user.ChangePasswordResult)
	}{
		{
			name:         "successful change",
			setupLimiter: passes,
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("old-password", "old-hashed-password").Return(true)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.Password == "new-hashed-password"
				})).Return(existing, nil)
			},
			setupRevoker: func(revoker *usermock.MockSessionRevoker) {
				revoker.EXPECT().DeleteSessionsByUserID(mock.Anything, userID, "current-session").Return(nil)
			},
			wantResult: func(t *testing.T, result *user.ChangePasswordResult) {
				assert.Equal(t, "password changed successfully", result.Message)
			},
		},
		{
			name:         "other sessions cannot be signed out",
			setupLimiter: passes,
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("old-password", "old-hashed-password").Return(true)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(existing, nil)
			},
			setupRevoker: func(revoker *usermock.MockSessionRevoker) {
				revoker.EXPECT().DeleteSessionsByUserID(mock.Anything, userID, "current-session").Return(errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "locked after repeated wrong passwords: refused before the password is checked",
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().RetryAfter(mock.Anything, key, auth.MaxFailedAttempts).Return(10*time.Minute, nil)
			},
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantErr: func(t *testing.T, err error) {
				var appErr *apperror.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, apperror.KindTooManyRequests, appErr.Kind)
				assert.Equal(t, 10*time.Minute, appErr.RetryAfter)
			},
		},
		{
			name: "failure that cannot be recorded is not turned into a plain rejection",
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "wrong-password",
				NewPassword: "NewPassword123",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter)
				limiter.EXPECT().RecordFailure(mock.Anything, key, auth.FailedAttemptsWindow).Return(errors.New("redis down"))
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("wrong-password", "old-hashed-password").Return(false)
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "limiter lookup failure is not ignored",
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().RetryAfter(mock.Anything, key, auth.MaxFailedAttempts).Return(time.Duration(0), errors.New("redis down"))
			},
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "validation error - empty old password and short new password",
			input: user.ChangePasswordInput{
				UserID:      uuid.Nil,
				SessionID:   "",
				OldPassword: "",
				NewPassword: "short",
			},
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "oldPassword")
				assert.Contains(t, validationErr.Fields, "newPassword")
			},
		},
		{
			name: "validation error - new password breaks complexity rules",
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "alllowercase",
			},
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.NotContains(t, validationErr.Fields, "oldPassword")
				assert.ElementsMatch(t, []string{
					"must contain at least one digit",
					"must contain at least one uppercase letter",
				}, validationErr.Fields["newPassword"])
			},
		},
		{
			name: "validation error - new password too long",
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "Aa1" + strings.Repeat("a", 30),
			},
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Equal(t, []string{"must be at most 32 characters"}, validationErr.Fields["newPassword"])
			},
		},
		{
			name: "validation error - cyrillic in new password",
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "Пароль123",
			},
			setupMocks: func(_ *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Equal(t, []string{"must not contain cyrillic characters"}, validationErr.Fields["newPassword"])
			},
		},
		{
			name:         "old password is not checked against complexity rules",
			setupLimiter: passes,
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "legacy",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("legacy", "old-hashed-password").Return(true)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(existing, nil)
			},
			setupRevoker: func(revoker *usermock.MockSessionRevoker) {
				revoker.EXPECT().DeleteSessionsByUserID(mock.Anything, userID, "current-session").Return(nil)
			},
			wantResult: func(t *testing.T, result *user.ChangePasswordResult) {
				assert.Equal(t, "password changed successfully", result.Message)
			},
		},
		{
			name:         "user not found",
			setupLimiter: notLocked,
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, _ *authmock.MockPasswordService) {
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrUserNotFound)
			},
		},
		{
			name: "invalid current password",
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter)
				limiter.EXPECT().RecordFailure(mock.Anything, key, auth.FailedAttemptsWindow).Return(nil)
			},
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "wrong-password",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("wrong-password", "old-hashed-password").Return(false)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, user.ErrInvalidCurrentPassword)
			},
		},
		{
			name:         "password hashing fails",
			setupLimiter: passes,
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("old-password", "old-hashed-password").Return(true)
				passwordService.EXPECT().Hash("NewPassword123").Return("", errors.New("bcrypt failure"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "bcrypt failure")
			},
		},
		{
			name:         "repository update error propagates",
			setupLimiter: passes,
			input: user.ChangePasswordInput{
				UserID:      userID,
				SessionID:   "current-session",
				OldPassword: "old-password",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, passwordService *authmock.MockPasswordService) {
				existing := &model.User{ID: userID, Password: "old-hashed-password"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				passwordService.EXPECT().Verify("old-password", "old-hashed-password").Return(true)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errors.New("update failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "update failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			passwordService := authmock.NewMockPasswordService(t)
			revoker := usermock.NewMockSessionRevoker(t)
			limiter := authmock.NewMockFailureLimiter(t)
			tt.setupMocks(userRepo, passwordService)
			if tt.setupRevoker != nil {
				tt.setupRevoker(revoker)
			}
			if tt.setupLimiter != nil {
				tt.setupLimiter(limiter)
			}

			service := user.NewService(userRepo, nil, passwordService, revoker, limiter)

			result, err := service.ChangePassword(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}
