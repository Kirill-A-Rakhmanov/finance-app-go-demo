package auth_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/auth"
	authmock "finance-app-go-demo/internal/application/auth/mock"
	apperror "finance-app-go-demo/internal/application/error"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_RefreshTokens(t *testing.T) {
	userID := uuid.New()
	tokenData := &auth.TokenData{UserID: userID, SessionID: "old-session"}

	tests := []struct {
		name       string
		input      auth.RefreshInput
		setupMocks func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *auth.RefreshResult)
	}{
		{
			name:  "successful refresh",
			input: auth.RefreshInput{RefreshToken: "valid-refresh-token"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("valid-refresh-token", auth.RefreshTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "old-session").Return(&auth.Session{ID: "old-session", UserID: userID}, nil)
				sessionStorage.EXPECT().CreateSession(mock.Anything, userID).Return(&auth.Session{ID: "new-session", UserID: userID}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(&auth.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"}, nil)
				sessionStorage.EXPECT().ShortenSession(mock.Anything, "old-session", auth.RefreshGracePeriod).Return(nil)
			},
			wantResult: func(t *testing.T, result *auth.RefreshResult) {
				assert.Equal(t, "new-access", result.TokenPair.AccessToken)
				assert.Equal(t, "new-refresh", result.TokenPair.RefreshToken)
			},
		},
		{
			name:  "validation error - empty refresh token",
			input: auth.RefreshInput{RefreshToken: ""},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "refreshToken")
			},
		},
		{
			name:  "invalid or expired token",
			input: auth.RefreshInput{RefreshToken: "garbage-token"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("garbage-token", auth.RefreshTokenType).Return(nil, errors.New("token expired"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "token expired")
			},
		},
		{
			name:  "issuing tokens fails: the old session stays untouched",
			input: auth.RefreshInput{RefreshToken: "valid-refresh-token"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("valid-refresh-token", auth.RefreshTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "old-session").Return(&auth.Session{ID: "old-session", UserID: userID}, nil)
				sessionStorage.EXPECT().CreateSession(mock.Anything, userID).Return(&auth.Session{ID: "new-session", UserID: userID}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(nil, errors.New("sign failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "sign failed")
			},
		},
		{
			name:  "retiring the old session fails",
			input: auth.RefreshInput{RefreshToken: "valid-refresh-token"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("valid-refresh-token", auth.RefreshTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "old-session").Return(&auth.Session{ID: "old-session", UserID: userID}, nil)
				sessionStorage.EXPECT().CreateSession(mock.Anything, userID).Return(&auth.Session{ID: "new-session", UserID: userID}, nil)
				tokenService.EXPECT().IssueTokens(mock.Anything).Return(&auth.TokenPair{AccessToken: "a", RefreshToken: "r"}, nil)
				sessionStorage.EXPECT().ShortenSession(mock.Anything, "old-session", auth.RefreshGracePeriod).Return(errors.New("redis down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis down")
			},
		},
		{
			name:  "session not found",
			input: auth.RefreshInput{RefreshToken: "valid-refresh-token"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				tokenService.EXPECT().VerifyToken("valid-refresh-token", auth.RefreshTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "old-session").Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrSessionNotFound)
			},
		},
		{
			name:  "session belongs to a different user",
			input: auth.RefreshInput{RefreshToken: "valid-refresh-token"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage, tokenService *authmock.MockTokenService) {
				otherUserID := uuid.New()
				tokenService.EXPECT().VerifyToken("valid-refresh-token", auth.RefreshTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "old-session").Return(&auth.Session{ID: "old-session", UserID: otherUserID}, nil)
				sessionStorage.EXPECT().DeleteSession(mock.Anything, "old-session").Return(nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrInvalidSession)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionStorage := authmock.NewMockSessionStorage(t)
			tokenService := authmock.NewMockTokenService(t)
			tt.setupMocks(sessionStorage, tokenService)

			service := auth.NewService(nil, nil, nil, tokenService, nil, nil, sessionStorage, nil, nil, nil, nil, nil)

			result, err := service.RefreshTokens(context.Background(), tt.input)

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
