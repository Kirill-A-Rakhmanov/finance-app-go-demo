package transaction_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_DeleteTransaction(t *testing.T) {
	transactionID := uuid.New()
	accountID := uuid.New()
	userID := uuid.New()
	transferID := uuid.New()

	tests := []struct {
		name       string
		input      transaction.DeleteTransactionInput
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction.DeleteTransactionResult)
	}{
		{
			name:  "successful deletion of unconfirmed transaction",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense, Amount: decimal.NewFromInt(30), IsConfirmed: false}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				transactionRepo.EXPECT().DeleteByID(mock.Anything, transactionID).Return(nil)
			},
			wantResult: func(t *testing.T, result *transaction.DeleteTransactionResult) {
				assert.True(t, decimal.NewFromInt(100).Equal(result.Balance))
			},
		},
		{
			name:  "successful deletion of confirmed expense adds amount back",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense, Amount: decimal.NewFromInt(30), IsConfirmed: true}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				transactionRepo.EXPECT().DeleteByID(mock.Anything, transactionID).Return(nil)
			},
			wantResult: func(t *testing.T, result *transaction.DeleteTransactionResult) {
				assert.True(t, decimal.NewFromInt(130).Equal(result.Balance))
			},
		},
		{
			name:  "successful deletion of confirmed income subtracts amount",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionIncome, Amount: decimal.NewFromInt(30), IsConfirmed: true}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, IsNegativeBalanceEnabled: true}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				transactionRepo.EXPECT().DeleteByID(mock.Anything, transactionID).Return(nil)
			},
			wantResult: func(t *testing.T, result *transaction.DeleteTransactionResult) {
				assert.True(t, decimal.NewFromInt(70).Equal(result.Balance))
			},
		},
		{
			name:  "validation error - empty fields",
			input: transaction.DeleteTransactionInput{TransactionID: uuid.Nil, UserID: uuid.Nil},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionId")
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "transaction not found",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction.ErrTransactionNotFound)
			},
		},
		{
			name:  "delete transfer not allowed",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense, TransferID: &transferID}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction.ErrDeleteTransferNotAllowed)
			},
		},
		{
			name:  "account not found",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name:  "negative balance not allowed after deletion",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionIncome, Amount: decimal.NewFromInt(30), IsConfirmed: true}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, IsNegativeBalanceEnabled: false}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(10), nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNegativeBalanceNotAllowed)
			},
		},
		{
			name:  "repository delete error",
			input: transaction.DeleteTransactionInput{TransactionID: transactionID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense, Amount: decimal.NewFromInt(30), IsConfirmed: false}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				transactionRepo.EXPECT().DeleteByID(mock.Anything, transactionID).Return(errors.New("delete failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "delete failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(accountRepo, transactionRepo)

			service := transaction.NewService(transactionRepo, accountRepo, nil, nil, newPassthroughTxManager(t), nil)

			result, err := service.DeleteTransaction(context.Background(), tt.input)

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
