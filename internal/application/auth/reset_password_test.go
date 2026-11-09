package auth_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ResetPassword(t *testing.T) {
	userID := uuid.New()
	email, err := valueObject.NewEmail("john@example.com")
	require.NoError(t, err)

	user := model.NewUser("John", "Doe", email, "old-hashed-password", nil)
	user.ID = userID

	tests := []struct {
		name       string
		input      auth.ResetPasswordInput
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			passwordService *authmock.MockPasswordService,
			recoveryStorage *authmock.MockRecoveryStorage,
		)
		setupSessions func(sessionStorage *authmock.MockSessionStorage)
		setupLimiter  func(limiter *authmock.MockFailureLimiter)
		wantErr       func(t *testing.T, err error)
		wantResult    func(t *testing.T, result *auth.ResetPasswordResult)
	}{
		{
			name: "successful reset",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.Password == "new-hashed-password"
				})).Return(user, nil)
				recoveryStorage.EXPECT().DeleteRecovery(mock.Anything, "recovery-code").Return(nil)
			},
			setupSessions: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSessionsByUserID(mock.Anything, userID, "").Return(nil)
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().Reset(mock.Anything, auth.LoginFailuresKey("john@example.com")).Return(nil)
			},
			wantResult: func(t *testing.T, result *auth.ResetPasswordResult) {
				assert.Equal(t, "password reset successfully", result.Message)
			},
		},
		{
			name: "validation error - empty code and short password",
			input: auth.ResetPasswordInput{
				Code:        "",
				NewPassword: "short",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRecoveryStorage,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "code")
				assert.Contains(t, validationErr.Fields, "newPassword")
			},
		},
		{
			name: "validation error - password missing digit and uppercase",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "lowercaseonly",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockRecoveryStorage,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				require.Contains(t, validationErr.Fields, "newPassword")
				assert.Contains(t, validationErr.Fields["newPassword"], "must contain at least one digit")
				assert.Contains(t, validationErr.Fields["newPassword"], "must contain at least one uppercase letter")
			},
		},
		{
			name: "recovery code not found",
			input: auth.ResetPasswordInput{
				Code:        "unknown-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "unknown-code").Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrInvalidRecoveryCode)
			},
		},
		{
			name: "user not found",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrInvalidRecoveryCode)
			},
		},
		{
			name: "password hashing fails",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil)
				passwordService.EXPECT().Hash("NewPassword123").Return("", errors.New("bcrypt failure"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "bcrypt failure")
			},
		},
		{
			name: "update fails",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errors.New("update failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "update failed")
			},
		},
		{
			name: "failing to lift the login lock keeps the code usable",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(user, nil)
			},
			setupSessions: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSessionsByUserID(mock.Anything, userID, "").Return(nil)
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().Reset(mock.Anything, auth.LoginFailuresKey("john@example.com")).Return(errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "session cleanup failure keeps the code usable",
			input: auth.ResetPasswordInput{
				Code:        "recovery-code",
				NewPassword: "NewPassword123",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				recoveryStorage *authmock.MockRecoveryStorage,
			) {
				recoveryStorage.EXPECT().GetRecovery(mock.Anything, "recovery-code").Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(user, nil)
				passwordService.EXPECT().Hash("NewPassword123").Return("new-hashed-password", nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(user, nil)
				// DeleteRecovery is not expected: the user can retry with the same code.
			},
			setupSessions: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSessionsByUserID(mock.Anything, userID, "").Return(errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			passwordService := authmock.NewMockPasswordService(t)
			recoveryStorage := authmock.NewMockRecoveryStorage(t)
			sessionStorage := authmock.NewMockSessionStorage(t)
			limiter := authmock.NewMockFailureLimiter(t)
			tt.setupMocks(userRepo, passwordService, recoveryStorage)
			if tt.setupSessions != nil {
				tt.setupSessions(sessionStorage)
			}
			if tt.setupLimiter != nil {
				tt.setupLimiter(limiter)
			}

			service := auth.NewService(userRepo, nil, nil, nil, passwordService, nil, sessionStorage, recoveryStorage, nil, nil, nil, limiter)

			result, err := service.ResetPassword(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}
