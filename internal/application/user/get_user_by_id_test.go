package user_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/user"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetUserById(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      user.GetUserByIdInput
		setupMocks func(userRepo *repomock.MockUserRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *user.GetUserByIdResult)
	}{
		{
			name:  "successful fetch",
			input: user.GetUserByIdInput{UserID: userID},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(&model.User{ID: userID, FirstName: "Jane"}, nil)
			},
			wantResult: func(t *testing.T, result *user.GetUserByIdResult) {
				assert.Equal(t, "Jane", result.User.FirstName)
			},
		},
		{
			name:  "validation error - empty user id",
			input: user.GetUserByIdInput{UserID: uuid.Nil},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "user not found",
			input: user.GetUserByIdInput{UserID: userID},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrUserNotFound)
			},
		},
		{
			name:  "repository error propagates",
			input: user.GetUserByIdInput{UserID: userID},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			tt.setupMocks(userRepo)

			service := user.NewService(userRepo, nil, nil, nil, nil)

			result, err := service.GetUserById(context.Background(), tt.input)

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
