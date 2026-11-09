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

func TestService_VerifyAuth(t *testing.T) {
	userID := uuid.New()
	tokenData := &auth.TokenData{UserID: userID, SessionID: "session-1"}

	tests := []struct {
		name       string
		input      auth.VerifyInput
		setupMocks func(
			sessionStorage *authmock.MockSessionStorage,
			tokenService *authmock.MockTokenService,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *auth.VerifyResult)
	}{
		{
			name:  "successful verification",
			input: auth.VerifyInput{AccessToken: "valid-access-token"},
			setupMocks: func(
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				tokenService.EXPECT().VerifyToken("valid-access-token", auth.AccessTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "session-1").Return(&auth.Session{ID: "session-1", UserID: userID}, nil)
			},
			wantResult: func(t *testing.T, result *auth.VerifyResult) {
				assert.Equal(t, userID, result.UserID)
				assert.Equal(t, "session-1", result.SessionID)
			},
		},
		{
			name:  "validation error - empty access token",
			input: auth.VerifyInput{AccessToken: ""},
			setupMocks: func(
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "accessToken")
			},
		},
		{
			name:  "invalid or expired token",
			input: auth.VerifyInput{AccessToken: "garbage-token"},
			setupMocks: func(
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				tokenService.EXPECT().VerifyToken("garbage-token", auth.AccessTokenType).Return(nil, errors.New("token expired"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "token expired")
			},
		},
		{
			name:  "session not found",
			input: auth.VerifyInput{AccessToken: "valid-access-token"},
			setupMocks: func(
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				tokenService.EXPECT().VerifyToken("valid-access-token", auth.AccessTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "session-1").Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, auth.ErrSessionNotFound)
			},
		},
		{
			name:  "session belongs to a different user",
			input: auth.VerifyInput{AccessToken: "valid-access-token"},
			setupMocks: func(
				sessionStorage *authmock.MockSessionStorage,
				tokenService *authmock.MockTokenService,
			) {
				otherUserID := uuid.New()
				tokenService.EXPECT().VerifyToken("valid-access-token", auth.AccessTokenType).Return(tokenData, nil)
				sessionStorage.EXPECT().GetSession(mock.Anything, "session-1").Return(&auth.Session{ID: "session-1", UserID: otherUserID}, nil)
				sessionStorage.EXPECT().DeleteSession(mock.Anything, "session-1").Return(nil)
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

			result, err := service.VerifyAuth(context.Background(), tt.input)

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
