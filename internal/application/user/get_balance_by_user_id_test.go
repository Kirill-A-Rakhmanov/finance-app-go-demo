package user_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/user"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetBalanceByUserID(t *testing.T) {
	userID := uuid.New()
	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	tests := []struct {
		name       string
		input      user.GetBalanceByUserIDInput
		setupMocks func(transactionRepo *repomock.MockTransactionRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *user.GetBalanceByUserIDResult)
	}{
		{
			name:  "successful fetch",
			input: user.GetBalanceByUserIDInput{UserID: userID},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetBalanceByUserID(mock.Anything, userID).Return([]valueObject.Balance{
					{Currency: usd, Amount: decimal.NewFromInt(150)},
				}, nil)
			},
			wantResult: func(t *testing.T, result *user.GetBalanceByUserIDResult) {
				require.Len(t, result.Balances, 1)
				assert.Equal(t, valueObject.CurrencyUSD, result.Balances[0].Currency.Code)
				assert.True(t, decimal.NewFromInt(150).Equal(result.Balances[0].Amount))
			},
		},
		{
			name:  "validation error - empty user id",
			input: user.GetBalanceByUserIDInput{UserID: uuid.Nil},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "repository error propagates",
			input: user.GetBalanceByUserIDInput{UserID: userID},
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetBalanceByUserID(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(transactionRepo)

			service := user.NewService(nil, transactionRepo, nil, nil, nil)

			result, err := service.GetBalanceByUserID(context.Background(), tt.input)

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
