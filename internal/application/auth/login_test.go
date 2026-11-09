package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestService_Login(t *testing.T) {
	userID := uuid.New()
	email, err := valueObject.NewEmail("john@example.com")
	require.NoError(t, err)

	user := model.NewUser("John", "Doe", email, "hashed-password", nil)
	user.ID = userID

	notLocked := func(limiter *authmock.MockFailureLimiter, key string) {
		limiter.EXPECT().RetryAfter(mock.Anything, key, auth.MaxFailedAttempts).Return(time.Duration(0), nil)
	}
	johnKey := auth.LoginFailuresKey("john@example.com")

	tests := []struct {
		name       string
		input      auth.LoginInput
		setupMocks func(
			userRepo *repomock.MockUserRepository,
			passwordService *authmock.MockPasswordService,
			sessionStorage *authmock.MockSessionStorage,
			tokenService *authmock.MockTokenService,
		)
		setupLimiter func(limiter *authmock.MockFailureLimiter)
		wantErr      func(t *testing.T, err error)
		wantResult   func(t *testing.T, result *auth.LoginResult)
	}{
		{
			name: "successful login",
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter, johnKey)
				limiter.EXPECT().Reset(mock.Anything, johnKey).Return(nil)
			},
			input: auth.LoginInput{
				Email:    "john@example.com",
				Password: "correct-password",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("correct-password", "hashed-password").Return(true)
				sessionStorage.EXPECT().CreateSession(mock.Anything, userID).Return(&auth.Session{ID: "session-1", UserID: userID}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(&auth.TokenPair{AccessToken: "access-token", RefreshToken: "refresh-token"}, nil)
			},
			wantResult: func(t *testing.T, result *auth.LoginResult) {
				assert.Equal(t, user, result.User)
				assert.Equal(t, "access-token", result.AccessToken)
				assert.Equal(t, "refresh-token", result.RefreshToken)
			},
		},
		{
			name: "email is normalized before the lookup",
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter, johnKey)
				limiter.EXPECT().Reset(mock.Anything, johnKey).Return(nil)
			},
			input: auth.LoginInput{
				Email:    "  John@Example.COM ",
				Password: "password",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("password", "hashed-password").Return(true)
				sessionStorage.EXPECT().CreateSession(mock.Anything, userID).Return(&auth.Session{ID: "session-1", UserID: userID}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(&auth.TokenPair{AccessToken: "a", RefreshToken: "r"}, nil)
			},
			wantResult: func(t *testing.T, result *auth.LoginResult) {
				assert.Equal(t, "a", result.AccessToken)
			},
		},
		{
			name: "user not found",
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				unknownKey := auth.LoginFailuresKey("unknown@example.com")
				notLocked(limiter, unknownKey)
				limiter.EXPECT().RecordFailure(mock.Anything, unknownKey, auth.FailedAttemptsWindow).Return(nil)
			},
			input: auth.LoginInput{
				Email:    "unknown@example.com",
				Password: "whatever",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "unknown@example.com").Return(nil, nil)
				// Verify still runs (with an empty hash) so the response time matches a wrong password.
				passwordService.EXPECT().Verify("whatever", "").Return(false)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
			},
		},
		{
			name: "invalid credentials",
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter, johnKey)
				limiter.EXPECT().RecordFailure(mock.Anything, johnKey, auth.FailedAttemptsWindow).Return(nil)
			},
			input: auth.LoginInput{
				Email:    "john@example.com",
				Password: "wrong-password",
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("wrong-password", "hashed-password").Return(false)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
			},
		},
		{
			name: "locked account is refused before anything else is checked",
			input: auth.LoginInput{
				Email:    "John@Example.com",
				Password: "whatever",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().RetryAfter(mock.Anything, johnKey, auth.MaxFailedAttempts).Return(4*time.Minute, nil)
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var appErr *apperror.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, apperror.KindTooManyRequests, appErr.Kind)
				assert.Equal(t, 4*time.Minute, appErr.RetryAfter)
			},
		},
		{
			name: "an unknown email is locked the same way as a known one",
			input: auth.LoginInput{
				Email:    "unknown@example.com",
				Password: "whatever",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().RetryAfter(mock.Anything, auth.LoginFailuresKey("unknown@example.com"), auth.MaxFailedAttempts).Return(time.Minute, nil)
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var appErr *apperror.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, apperror.KindTooManyRequests, appErr.Kind)
			},
		},
		{
			name: "limiter lookup failure is not ignored",
			input: auth.LoginInput{
				Email:    "john@example.com",
				Password: "correct-password",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				limiter.EXPECT().RetryAfter(mock.Anything, johnKey, auth.MaxFailedAttempts).Return(time.Duration(0), errors.New("redis down"))
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "failure that cannot be recorded is not turned into a plain rejection",
			input: auth.LoginInput{
				Email:    "john@example.com",
				Password: "wrong-password",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter, johnKey)
				limiter.EXPECT().RecordFailure(mock.Anything, johnKey, auth.FailedAttemptsWindow).Return(errors.New("redis down"))
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("wrong-password", "hashed-password").Return(false)
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "counter reset failure stops the login before a session is created",
			input: auth.LoginInput{
				Email:    "john@example.com",
				Password: "correct-password",
			},
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter, johnKey)
				limiter.EXPECT().Reset(mock.Anything, johnKey).Return(errors.New("redis down"))
			},
			setupMocks: func(
				userRepo *repomock.MockUserRepository,
				passwordService *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(user, nil)
				passwordService.EXPECT().Verify("correct-password", "hashed-password").Return(true)
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name: "validation error - empty fields",
			input: auth.LoginInput{
				Email:    "",
				Password: "",
			},
			setupMocks: func(
				_ *repomock.MockUserRepository,
				_ *authmock.MockPasswordService,
				_ *authmock.MockSessionStorage,
				_ *authmock.MockTokenService,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "email")
				assert.Contains(t, validationErr.Fields, "password")
			},
		},
		{
			name: "repository error propagates",
			setupLimiter: func(limiter *authmock.MockFailureLimiter) {
				notLocked(limiter, johnKey)
			},
			input: auth.LoginInput{
				Email:    "john@example.com",
				Password: "correct-password",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository, _ *authmock.MockPasswordService, _ *authmock.MockSessionStorage, _ *authmock.MockTokenService) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
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
			tt.setupMocks(userRepo, passwordService, sessionStorage, tokenService)
			if tt.setupLimiter != nil {
				tt.setupLimiter(limiter)
			}

			service := auth.NewService(userRepo, nil, nil, tokenService, passwordService, nil, sessionStorage, nil, nil, nil, nil, limiter)

			result, err := service.Login(context.Background(), tt.input)

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
