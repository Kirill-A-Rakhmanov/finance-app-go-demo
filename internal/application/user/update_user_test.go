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

func TestService_UpdateUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      user.UpdateUserInput
		setupMocks func(userRepo *repomock.MockUserRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *user.UpdateUserResult)
	}{
		{
			name: "successful update",
			input: user.UpdateUserInput{
				UserID:    userID,
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				existing := &model.User{ID: userID, FirstName: "John", LastName: "Doe"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
				userRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.FirstName == "Jane" && u.LastName == "Smith"
				})).Return(&model.User{ID: userID, FirstName: "Jane", LastName: "Smith"}, nil)
			},
			wantResult: func(t *testing.T, result *user.UpdateUserResult) {
				assert.Equal(t, "Jane", result.User.FirstName)
				assert.Equal(t, "Smith", result.User.LastName)
			},
		},
		{
			name: "validation error - empty fields",
			input: user.UpdateUserInput{
				UserID:    uuid.Nil,
				FirstName: "",
				LastName:  "",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "firstName")
				assert.Contains(t, validationErr.Fields, "lastName")
			},
		},
		{
			name: "user not found",
			input: user.UpdateUserInput{
				UserID:    userID,
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrUserNotFound)
			},
		},
		{
			name: "repository update error propagates",
			input: user.UpdateUserInput{
				UserID:    userID,
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				existing := &model.User{ID: userID, FirstName: "John", LastName: "Doe"}
				userRepo.EXPECT().GetByID(mock.Anything, userID).Return(existing, nil)
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
			tt.setupMocks(userRepo)

			service := user.NewService(userRepo, nil, nil, nil, nil)

			result, err := service.UpdateUser(context.Background(), tt.input)

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
