package transfer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transfer"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func runTx(txManager *transactionmock.MockManager) {
	txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		})
	txManager.EXPECT().Lock(mock.Anything, mock.Anything).Return(nil).Maybe()
}

func TestService_CreateTransfer(t *testing.T) {
	userID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)
	eur, err := valueObject.NewCurrency(valueObject.CurrencyEUR)
	require.NoError(t, err)

	baseInput := transfer.CreateTransferInput{
		UserID:        userID,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		FromAmount:    decimal.NewFromInt(50),
		ToAmount:      decimal.NewFromInt(50),
		Date:          time.Now(),
	}

	tests := []struct {
		name       string
		input      transfer.CreateTransferInput
		setupMocks func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transfer.CreateTransferResult)
	}{
		{
			name:  "successful transfer same currency",
			input: baseInput,
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, Currency: usd}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, fromAccountID).Return(decimal.NewFromInt(100), nil)
				transferRepo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tr *model.Transfer) (*model.Transfer, error) {
						tr.ID = uuid.New()
						return tr, nil
					})
				transactionRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, txs []model.Transaction) ([]model.Transaction, error) {
						return txs, nil
					})
			},
			wantResult: func(t *testing.T, result *transfer.CreateTransferResult) {
				assert.Len(t, result.Transactions, 2)
			},
		},
		{
			name: "successful transfer different currency",
			input: transfer.CreateTransferInput{
				UserID:        userID,
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				FromAmount:    decimal.NewFromInt(100),
				ToAmount:      decimal.NewFromInt(90),
				Date:          time.Now(),
			},
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, Currency: eur}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, fromAccountID).Return(decimal.NewFromInt(200), nil)
				transferRepo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tr *model.Transfer) (*model.Transfer, error) {
						tr.ID = uuid.New()
						return tr, nil
					})
				transactionRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, txs []model.Transaction) ([]model.Transaction, error) {
						return txs, nil
					})
			},
			wantResult: func(t *testing.T, result *transfer.CreateTransferResult) {
				assert.True(t, decimal.NewFromInt(100).Equal(result.Transfer.FromAmount))
				assert.True(t, decimal.NewFromInt(90).Equal(result.Transfer.ToAmount))
			},
		},
		{
			name: "validation error - empty accounts and zero amounts",
			input: transfer.CreateTransferInput{
				UserID:        userID,
				FromAccountID: uuid.Nil,
				ToAccountID:   uuid.Nil,
				FromAmount:    decimal.Zero,
				ToAmount:      decimal.Zero,
			},
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				// no calls expected - Validate() must short-circuit before touching any dependency
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "fromAccountId")
				assert.Contains(t, validationErr.Fields, "toAccountId")
				assert.Contains(t, validationErr.Fields, "fromAmount")
				assert.Contains(t, validationErr.Fields, "toAmount")
			},
		},
		{
			name: "transfer to same account not allowed",
			input: transfer.CreateTransferInput{
				UserID:        userID,
				FromAccountID: fromAccountID,
				ToAccountID:   fromAccountID,
				FromAmount:    decimal.NewFromInt(50),
				ToAmount:      decimal.NewFromInt(50),
			},
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				// no calls expected - checked before entering the transaction
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transfer.ErrTransferToSameAccountNotAllowed)
			},
		},
		{
			name:  "from account not found",
			input: baseInput,
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name:  "to account not found",
			input: baseInput,
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name: "amount mismatch for same currency",
			input: transfer.CreateTransferInput{
				UserID:        userID,
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				FromAmount:    decimal.NewFromInt(50),
				ToAmount:      decimal.NewFromInt(40),
			},
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, Currency: usd}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transfer.ErrAmountMismatchForSameCurrency)
			},
		},
		{
			name:  "negative balance not allowed",
			input: baseInput,
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd, IsNegativeBalanceEnabled: false}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, Currency: usd}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, fromAccountID).Return(decimal.NewFromInt(10), nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNegativeBalanceNotAllowed)
			},
		},
		{
			name:  "negative balance allowed skips balance check",
			input: baseInput,
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd, IsNegativeBalanceEnabled: true}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, Currency: usd}, nil)
				transferRepo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tr *model.Transfer) (*model.Transfer, error) {
						tr.ID = uuid.New()
						return tr, nil
					})
				transactionRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, txs []model.Transaction) ([]model.Transaction, error) {
						return txs, nil
					})
			},
			wantResult: func(t *testing.T, result *transfer.CreateTransferResult) {
				assert.Len(t, result.Transactions, 2)
			},
		},
		{
			name:  "repository create error",
			input: baseInput,
			setupMocks: func(accountRepo *repomock.MockAccountRepository, transactionRepo *repomock.MockTransactionRepository, transferRepo *repomock.MockTransferRepository, txManager *transactionmock.MockManager) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, userID).
					Return(&model.Account{ID: fromAccountID, UserID: userID, Currency: usd}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, userID).
					Return(&model.Account{ID: toAccountID, UserID: userID, Currency: usd}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, fromAccountID).Return(decimal.NewFromInt(100), nil)
				transferRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			transferRepo := repomock.NewMockTransferRepository(t)
			txManager := transactionmock.NewMockManager(t)
			tt.setupMocks(accountRepo, transactionRepo, transferRepo, txManager)

			service := transfer.NewService(transferRepo, transactionRepo, accountRepo, txManager)

			result, err := service.CreateTransfer(context.Background(), tt.input)

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
