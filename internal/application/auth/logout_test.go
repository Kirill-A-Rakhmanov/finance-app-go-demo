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

func TestService_Logout(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      auth.LogoutInput
		setupMocks func(sessionStorage *authmock.MockSessionStorage)
		wantErr    func(t *testing.T, err error)
	}{
		{
			name:  "successful logout",
			input: auth.LogoutInput{UserID: userID, SessionID: "session-1"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSession(mock.Anything, "session-1").Return(nil)
			},
		},
		{
			name:  "validation error - empty fields",
			input: auth.LogoutInput{UserID: uuid.Nil, SessionID: ""},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "sessionId")
			},
		},
		{
			name:  "session storage error propagates",
			input: auth.LogoutInput{UserID: userID, SessionID: "session-1"},
			setupMocks: func(sessionStorage *authmock.MockSessionStorage) {
				sessionStorage.EXPECT().DeleteSession(mock.Anything, "session-1").Return(errors.New("redis unavailable"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "redis unavailable")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionStorage := authmock.NewMockSessionStorage(t)
			tt.setupMocks(sessionStorage)

			service := auth.NewService(nil, nil, nil, nil, nil, nil, sessionStorage, nil, nil, nil, nil, nil)

			err := service.Logout(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
