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

func TestService_Recovery(t *testing.T) {
	userID := uuid.New()
	email, err := valueObject.NewEmail("john@example.com")
	require.NoError(t, err)

	user := model.NewUser("John", "Doe", email, "hashed-password", nil)
	user.ID = userID

	tests := []struct {
		name       string
		input      auth.RecoveryInput
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			recoveryStorage *authmock.MockRecoveryStorage,
			recoveryNotifier *authmock.MockRecoveryNotifier,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *auth.RecoveryResult)
	}{
		{
			name: "successful recovery request",
			input: auth.RecoveryInput{
				Email: "john@example.com",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				recoveryNotifier *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "john@example.com", auth.RecoveryCooldown).Return(true, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				recoveryStorage.EXPECT().CreateRecovery(mock.Anything, userID).Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				recoveryNotifier.EXPECT().SendPasswordRecovery(mock.Anything, "john@example.com", "recovery-code").Return(nil)
			},
			wantResult: func(t *testing.T, result *auth.RecoveryResult) {
				assert.Equal(t, "if an account with this email exists, a recovery code has been sent", result.Message)
			},
		},
		{
			name: "email is normalized before the lookup",
			input: auth.RecoveryInput{
				Email: " John@Example.COM",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				recoveryNotifier *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "john@example.com", auth.RecoveryCooldown).Return(true, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				recoveryStorage.EXPECT().CreateRecovery(mock.Anything, userID).Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				recoveryNotifier.EXPECT().SendPasswordRecovery(mock.Anything, "john@example.com", "recovery-code").Return(nil)
			},
			wantResult: func(t *testing.T, result *auth.RecoveryResult) {
				assert.NotEmpty(t, result.Message)
			},
		},
		{
			name: "within the cooldown: same answer, nothing is looked up, created or sent",
			input: auth.RecoveryInput{
				Email: "john@example.com",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "john@example.com", auth.RecoveryCooldown).Return(false, nil)
			},
			wantResult: func(t *testing.T, result *auth.RecoveryResult) {
				assert.Equal(t, "if an account with this email exists, a recovery code has been sent", result.Message)
			},
		},
		{
			name: "cooldown storage error propagates",
			input: auth.RecoveryInput{
				Email: "john@example.com",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "john@example.com", auth.RecoveryCooldown).Return(false, errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "validation error - empty email",
			input: auth.RecoveryInput{
				Email: "",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "email")
			},
		},
		{
			name: "user not found",
			input: auth.RecoveryInput{
				Email: "unknown@example.com",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "unknown@example.com", auth.RecoveryCooldown).Return(true, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "unknown@example.com").Return(nil, nil)
			},
			// Same answer as for an existing account, and no code is created or sent
			// (unexpected storage/notifier calls would fail the mocks).
			wantResult: func(t *testing.T, result *auth.RecoveryResult) {
				assert.Equal(t, "if an account with this email exists, a recovery code has been sent", result.Message)
			},
		},
		{
			name: "recovery storage error propagates",
			input: auth.RecoveryInput{
				Email: "john@example.com",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				_ *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "john@example.com", auth.RecoveryCooldown).Return(true, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				recoveryStorage.EXPECT().CreateRecovery(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
		{
			name: "notifier error propagates",
			input: auth.RecoveryInput{
				Email: "john@example.com",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				recoveryStorage *authmock.MockRecoveryStorage,
				recoveryNotifier *authmock.MockRecoveryNotifier,
			) {
				recoveryStorage.EXPECT().AcquireCooldown(mock.Anything, "john@example.com", auth.RecoveryCooldown).Return(true, nil)
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				recoveryStorage.EXPECT().CreateRecovery(mock.Anything, userID).Return(&auth.Recovery{Code: "recovery-code", UserID: userID}, nil)
				recoveryNotifier.EXPECT().SendPasswordRecovery(mock.Anything, "john@example.com", "recovery-code").Return(errors.New("send failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "send failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			recoveryStorage := authmock.NewMockRecoveryStorage(t)
			recoveryNotifier := authmock.NewMockRecoveryNotifier(t)
			tt.setupMocks(userRepo, recoveryStorage, recoveryNotifier)

			service := auth.NewService(userRepo, nil, nil, nil, nil, nil, nil, recoveryStorage, recoveryNotifier, nil, nil, nil)

			result, err := service.Recovery(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}
