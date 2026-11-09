package transfer_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transfer"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_DeleteTransfer(t *testing.T) {
	transferID := uuid.New()
	toAccountID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      transfer.DeleteTransferInput
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
			transferRepo *repomock.MockTransferRepository,
		)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transfer.DeleteTransferResult)
	}{
		{
			name:  "successful deletion",
			input: transfer.DeleteTransferInput{TransferID: transferID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).
					Return(&model.Transfer{ID: transferID, ToAccountID: toAccountID, ToAmount: decimal.NewFromInt(50)}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, IsNegativeBalanceEnabled: true}, nil)
				transferRepo.EXPECT().DeleteByID(mock.Anything, transferID).Return(nil)
			},
			wantResult: func(t *testing.T, result *transfer.DeleteTransferResult) {
				assert.Equal(t, "Transfer deleted successfully", result.Message)
			},
		},
		{
			name:  "transfer not found",
			input: transfer.DeleteTransferInput{TransferID: transferID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transfer.ErrTransferNotFound)
			},
		},
		{
			name:  "to account not found",
			input: transfer.DeleteTransferInput{TransferID: transferID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).
					Return(&model.Transfer{ID: transferID, ToAccountID: toAccountID}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name:  "negative balance not allowed",
			input: transfer.DeleteTransferInput{TransferID: transferID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).
					Return(&model.Transfer{ID: transferID, ToAccountID: toAccountID, ToAmount: decimal.NewFromInt(50)}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, IsNegativeBalanceEnabled: false}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, toAccountID).Return(decimal.NewFromInt(20), nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNegativeBalanceNotAllowed)
			},
		},
		{
			name:  "repository delete error",
			input: transfer.DeleteTransferInput{TransferID: transferID, UserID: userID},
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).
					Return(&model.Transfer{ID: transferID, ToAccountID: toAccountID, ToAmount: decimal.NewFromInt(50)}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, IsNegativeBalanceEnabled: true}, nil)
				transferRepo.EXPECT().DeleteByID(mock.Anything, transferID).Return(errors.New("delete failed"))
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
			transferRepo := repomock.NewMockTransferRepository(t)
			txManager := transactionmock.NewMockManager(t)
			txManager.EXPECT().
				Wrap(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
					return txFunc(ctx)
				}).
				Maybe()
			txManager.EXPECT().Lock(mock.Anything, mock.Anything).Return(nil).Maybe()
			tt.setupMocks(accountRepo, transactionRepo, transferRepo)

			service := transfer.NewService(transferRepo, transactionRepo, accountRepo, txManager)

			result, err := service.DeleteTransfer(context.Background(), tt.input)

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
