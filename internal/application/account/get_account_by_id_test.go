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

func TestService_GetAccountById(t *testing.T) {
	accountID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      account.GetAccountByIdInput
		setupMocks func(accountRepo *repomock.MockAccountRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *account.GetAccountByIdResult)
	}{
		{
			name:  "successful fetch",
			input: account.GetAccountByIdInput{AccountID: accountID, UserID: userID},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, Name: "My Account"}, nil)
			},
			wantResult: func(t *testing.T, result *account.GetAccountByIdResult) {
				assert.Equal(t, "My Account", result.Account.Name)
			},
		},
		{
			name:  "validation error - empty fields",
			input: account.GetAccountByIdInput{AccountID: uuid.Nil, UserID: uuid.Nil},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "accountId")
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "account not found",
			input: account.GetAccountByIdInput{AccountID: accountID, UserID: userID},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name:  "repository error propagates",
			input: account.GetAccountByIdInput{AccountID: accountID, UserID: userID},
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, errors.New("connection refused"))
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

			result, err := service.GetAccountById(context.Background(), tt.input)

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
