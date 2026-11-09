package transaction_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestService_CreateExpenseTransaction(t *testing.T) {
	accountID := uuid.New()
	userID := uuid.New()
	existingCategoryID := uuid.New()
	existingCounterpartyID := uuid.New()
	createdCategoryID := uuid.New()
	createdCounterpartyID := uuid.New()

	categoryName := "Groceries"
	counterpartyName := "Supermarket"
	blankName := "   "
	tooLongCategoryName := strings.Repeat("a", model.TransactionCategoryNameMaxLen+1)
	tooLongCounterpartyName := strings.Repeat("a", model.CounterpartyNameMaxLen+1)

	type mocks struct {
		account      *repomock.MockAccountRepository
		transaction  *repomock.MockTransactionRepository
		category     *repomock.MockTransactionCategoryRepository
		counterparty *repomock.MockCounterpartyRepository
	}

	expectCreate := func(m mocks) {
		m.transaction.EXPECT().Create(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
				tx.ID = uuid.New()
				return tx, nil
			})
	}

	tests := []struct {
		name       string
		input      transaction.CreateExpenseTransactionInput
		setupMocks func(m mocks)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction.CreateExpenseTransactionResult)
	}{
		{
			name: "successful creation without category and counterparty",
			input: transaction.CreateExpenseTransactionInput{
				UserID:    userID,
				AccountID: accountID,
				Amount:    decimal.NewFromInt(50),
				Date:      time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				expectCreate(m)
			},
			wantResult: func(t *testing.T, result *transaction.CreateExpenseTransactionResult) {
				assert.Equal(t, enum.TransactionExpense, result.Transaction.Type)
				assert.True(t, decimal.NewFromInt(50).Equal(result.Transaction.Amount))
				assert.Nil(t, result.Transaction.TransactionCategoryID)
				assert.Nil(t, result.Transaction.CounterpartyID)
			},
		},
		{
			name: "blank names are treated as not provided",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &blankName,
				CounterpartyName:        &blankName,
				Amount:                  decimal.NewFromInt(50),
				Date:                    time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				expectCreate(m)
			},
			wantResult: func(t *testing.T, result *transaction.CreateExpenseTransactionResult) {
				assert.Nil(t, result.Transaction.TransactionCategoryID)
				assert.Nil(t, result.Transaction.CounterpartyID)
			},
		},
		{
			name: "uses the existing category and counterparty",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &categoryName,
				CounterpartyName:        &counterpartyName,
				Amount:                  decimal.NewFromInt(50),
				Date:                    time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).
					Return(&model.TransactionCategory{ID: existingCategoryID, UserID: userID, Name: categoryName, Type: enum.Expense}, nil)
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).
					Return(&model.Counterparty{ID: existingCounterpartyID, UserID: userID, Name: counterpartyName}, nil)
				expectCreate(m)
			},
			wantResult: func(t *testing.T, result *transaction.CreateExpenseTransactionResult) {
				require.NotNil(t, result.Transaction.TransactionCategoryID)
				assert.Equal(t, existingCategoryID, *result.Transaction.TransactionCategoryID)
				require.NotNil(t, result.Transaction.CounterpartyID)
				assert.Equal(t, existingCounterpartyID, *result.Transaction.CounterpartyID)
			},
		},
		{
			name: "creates the missing category and counterparty",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &categoryName,
				CounterpartyName:        &counterpartyName,
				Amount:                  decimal.NewFromInt(50),
				Date:                    time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).Return(nil, nil)
				m.category.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						assert.Equal(t, userID, c.UserID)
						assert.Equal(t, categoryName, c.Name)
						assert.Equal(t, enum.Expense, c.Type, "category type follows the transaction type")
						assert.Nil(t, c.ParentID)
						assert.Equal(t, model.DefaultTransactionCategoryIcon, c.Icon)
						assert.Equal(t, model.DefaultTransactionCategoryColor, c.Color)
						c.ID = createdCategoryID
						return c, nil
					})
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).Return(nil, nil)
				m.counterparty.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.Counterparty) (*model.Counterparty, error) {
						assert.Equal(t, userID, c.UserID)
						assert.Equal(t, counterpartyName, c.Name)
						c.ID = createdCounterpartyID
						return c, nil
					})
				expectCreate(m)
			},
			wantResult: func(t *testing.T, result *transaction.CreateExpenseTransactionResult) {
				require.NotNil(t, result.Transaction.TransactionCategoryID)
				assert.Equal(t, createdCategoryID, *result.Transaction.TransactionCategoryID)
				require.NotNil(t, result.Transaction.CounterpartyID)
				assert.Equal(t, createdCounterpartyID, *result.Transaction.CounterpartyID)
			},
		},
		{
			name: "names are trimmed before lookup",
			input: func() transaction.CreateExpenseTransactionInput {
				padded := "  " + categoryName + " "
				return transaction.CreateExpenseTransactionInput{
					UserID:                  userID,
					AccountID:               accountID,
					TransactionCategoryName: &padded,
					Amount:                  decimal.NewFromInt(50),
					Date:                    time.Now(),
				}
			}(),
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).
					Return(&model.TransactionCategory{ID: existingCategoryID, UserID: userID, Name: categoryName, Type: enum.Expense}, nil)
				expectCreate(m)
			},
			wantResult: func(t *testing.T, result *transaction.CreateExpenseTransactionResult) {
				require.NotNil(t, result.Transaction.TransactionCategoryID)
				assert.Equal(t, existingCategoryID, *result.Transaction.TransactionCategoryID)
			},
		},
		{
			name: "negative balance not allowed on account without overdraft",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &categoryName,
				CounterpartyName:        &counterpartyName,
				Amount:                  decimal.NewFromInt(150),
				Date:                    time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, IsNegativeBalanceEnabled: false}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				// no category/counterparty calls: nothing may be created for a transaction that is rejected
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNegativeBalanceNotAllowed)
			},
		},
		{
			name: "negative balance allowed skips balance check",
			input: transaction.CreateExpenseTransactionInput{
				UserID:    userID,
				AccountID: accountID,
				Amount:    decimal.NewFromInt(150),
				Date:      time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, IsNegativeBalanceEnabled: true}, nil)
				expectCreate(m)
			},
			wantResult: func(t *testing.T, result *transaction.CreateExpenseTransactionResult) {
				assert.True(t, decimal.NewFromInt(150).Equal(result.Transaction.Amount))
			},
		},
		{
			name: "validation error - empty account and zero amount",
			input: transaction.CreateExpenseTransactionInput{
				UserID:    userID,
				AccountID: uuid.Nil,
				Amount:    decimal.Zero,
			},
			setupMocks: func(m mocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "accountId")
				assert.Contains(t, validationErr.Fields, "amount")
			},
		},
		{
			name: "validation error - names too long",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &tooLongCategoryName,
				CounterpartyName:        &tooLongCounterpartyName,
				Amount:                  decimal.NewFromInt(50),
			},
			setupMocks: func(m mocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionCategoryName")
				assert.Contains(t, validationErr.Fields, "counterpartyName")
			},
		},
		{
			name: "account not found",
			input: transaction.CreateExpenseTransactionInput{
				UserID:    userID,
				AccountID: accountID,
				Amount:    decimal.NewFromInt(50),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name: "category lookup error",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &categoryName,
				Amount:                  decimal.NewFromInt(50),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).
					Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name: "category create error",
			input: transaction.CreateExpenseTransactionInput{
				UserID:                  userID,
				AccountID:               accountID,
				TransactionCategoryName: &categoryName,
				Amount:                  decimal.NewFromInt(50),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).Return(nil, nil)
				m.category.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
		{
			name: "counterparty lookup error",
			input: transaction.CreateExpenseTransactionInput{
				UserID:           userID,
				AccountID:        accountID,
				CounterpartyName: &counterpartyName,
				Amount:           decimal.NewFromInt(50),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).
					Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name: "counterparty create error",
			input: transaction.CreateExpenseTransactionInput{
				UserID:           userID,
				AccountID:        accountID,
				CounterpartyName: &counterpartyName,
				Amount:           decimal.NewFromInt(50),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).Return(nil, nil)
				m.counterparty.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
		{
			name: "repository create error",
			input: transaction.CreateExpenseTransactionInput{
				UserID:    userID,
				AccountID: accountID,
				Amount:    decimal.NewFromInt(50),
				Date:      time.Now(),
			},
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				m.transaction.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks{
				account:      repomock.NewMockAccountRepository(t),
				transaction:  repomock.NewMockTransactionRepository(t),
				category:     repomock.NewMockTransactionCategoryRepository(t),
				counterparty: repomock.NewMockCounterpartyRepository(t),
			}
			tt.setupMocks(m)

			service := transaction.NewService(m.transaction, m.account, m.category, m.counterparty, newPassthroughTxManager(t), nil)

			result, err := service.CreateExpenseTransaction(context.Background(), tt.input)

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
