package account_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/account"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetAccountsByUserId(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      account.GetAccountsByUserIdInput
		setupMocks func(accountRepo *repomock.MockAccountRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *account.GetAccountsByUserIdResult)
	}{
		{
			name:  "successful fetch",
			input: account.GetAccountsByUserIdInput{UserID: userID},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.Account{
					{ID: uuid.New(), UserID: userID, Name: "Cash"},
					{ID: uuid.New(), UserID: userID, Name: "Card"},
				}, nil)
			},
			wantResult: func(t *testing.T, result *account.GetAccountsByUserIdResult) {
				assert.Len(t, result.Accounts, 2)
			},
		},
		{
			name:  "validation error - empty user id",
			input: account.GetAccountsByUserIdInput{UserID: uuid.Nil},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "repository error propagates",
			input: account.GetAccountsByUserIdInput{UserID: userID},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByUserID(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			tt.setupMocks(accountRepo)

			service := account.NewService(nil, accountRepo, nil, nil)

			result, err := service.GetAccountsByUserId(context.Background(), tt.input)

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
