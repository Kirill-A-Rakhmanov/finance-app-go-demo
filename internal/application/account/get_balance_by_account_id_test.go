package account_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/account"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetBalanceByAccountId(t *testing.T) {
	accountID := uuid.New()
	userID := uuid.New()
	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	tests := []struct {
		name       string
		input      account.GetBalanceByAccountIdInput
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *account.GetBalanceByAccountIdResult)
	}{
		{
			name:  "successful fetch",
			input: account.GetBalanceByAccountIdInput{AccountID: accountID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, Currency: usd}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(150), nil)
			},
			wantResult: func(t *testing.T, result *account.GetBalanceByAccountIdResult) {
				assert.True(t, decimal.NewFromInt(150).Equal(result.Balance))
				assert.Equal(t, valueObject.CurrencyUSD, result.Currency.Code)
			},
		},
		{
			name:  "validation error - empty fields",
			input: account.GetBalanceByAccountIdInput{AccountID: uuid.Nil, UserID: uuid.Nil},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
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
			input: account.GetBalanceByAccountIdInput{AccountID: accountID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name:  "balance repository error propagates",
			input: account.GetBalanceByAccountIdInput{AccountID: accountID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, Currency: usd}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.Decimal{}, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(accountRepo, transactionRepo)

			service := account.NewService(nil, accountRepo, transactionRepo, nil)

			result, err := service.GetBalanceByAccountId(context.Background(), tt.input)

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
