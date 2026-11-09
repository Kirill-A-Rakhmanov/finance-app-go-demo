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

func TestService_UpdateTransaction(t *testing.T) {
	transactionID := uuid.New()
	accountID := uuid.New()
	userID := uuid.New()
	transferID := uuid.New()
	existingCategoryID := uuid.New()
	existingCounterpartyID := uuid.New()
	createdCategoryID := uuid.New()
	createdCounterpartyID := uuid.New()

	categoryName := "Groceries"
	counterpartyName := "Supermarket"
	blankName := "   "
	tooLongCategoryName := strings.Repeat("a", model.TransactionCategoryNameMaxLen+1)
	tooLongCounterpartyName := strings.Repeat("a", model.CounterpartyNameMaxLen+1)

	baseInput := transaction.UpdateTransactionInput{
		TransactionID: transactionID,
		UserID:        userID,
		AccountID:     accountID,
		Amount:        decimal.NewFromInt(50),
		Type:          enum.TransactionExpense,
		Date:          time.Now(),
	}
	withNames := func(category, counterparty *string) transaction.UpdateTransactionInput {
		i := baseInput
		i.TransactionCategoryName = category
		i.CounterpartyName = counterparty
		return i
	}

	type mocks struct {
		account      *repomock.MockAccountRepository
		transaction  *repomock.MockTransactionRepository
		category     *repomock.MockTransactionCategoryRepository
		counterparty *repomock.MockCounterpartyRepository
	}

	// expectLoaded mocks the lookups; the balance is not read because these updates do not lower it.
	expectLoaded := func(m mocks, stored model.Transaction) {
		m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
			Return(&model.Account{ID: accountID, UserID: userID}, nil)
		m.transaction.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(&stored, nil)
	}
	storedExpense := model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense}

	expectUpdate := func(m mocks) {
		m.transaction.EXPECT().Update(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
				return tx, nil
			})
	}

	tests := []struct {
		name       string
		input      transaction.UpdateTransactionInput
		setupMocks func(m mocks)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction.UpdateTransactionResult)
	}{
		{
			name:  "successful update",
			input: baseInput,
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				expectUpdate(m)
			},
			wantResult: func(t *testing.T, result *transaction.UpdateTransactionResult) {
				assert.True(t, decimal.NewFromInt(50).Equal(result.Transaction.Amount))
			},
		},
		{
			name:  "no names clear the existing category and counterparty",
			input: baseInput,
			setupMocks: func(m mocks) {
				expectLoaded(m, model.Transaction{
					ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense,
					TransactionCategoryID: &existingCategoryID, CounterpartyID: &existingCounterpartyID,
				})
				expectUpdate(m)
			},
			wantResult: func(t *testing.T, result *transaction.UpdateTransactionResult) {
				assert.Nil(t, result.Transaction.TransactionCategoryID)
				assert.Nil(t, result.Transaction.CounterpartyID)
			},
		},
		{
			name:  "blank names clear the existing category and counterparty",
			input: withNames(&blankName, &blankName),
			setupMocks: func(m mocks) {
				expectLoaded(m, model.Transaction{
					ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense,
					TransactionCategoryID: &existingCategoryID, CounterpartyID: &existingCounterpartyID,
				})
				expectUpdate(m)
			},
			wantResult: func(t *testing.T, result *transaction.UpdateTransactionResult) {
				assert.Nil(t, result.Transaction.TransactionCategoryID)
				assert.Nil(t, result.Transaction.CounterpartyID)
			},
		},
		{
			name:  "uses the existing category and counterparty",
			input: withNames(&categoryName, &counterpartyName),
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).
					Return(&model.TransactionCategory{ID: existingCategoryID, UserID: userID, Name: categoryName, Type: enum.Expense}, nil)
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).
					Return(&model.Counterparty{ID: existingCounterpartyID, UserID: userID, Name: counterpartyName}, nil)
				expectUpdate(m)
			},
			wantResult: func(t *testing.T, result *transaction.UpdateTransactionResult) {
				require.NotNil(t, result.Transaction.TransactionCategoryID)
				assert.Equal(t, existingCategoryID, *result.Transaction.TransactionCategoryID)
				require.NotNil(t, result.Transaction.CounterpartyID)
				assert.Equal(t, existingCounterpartyID, *result.Transaction.CounterpartyID)
			},
		},
		{
			name:  "creates the missing category and counterparty",
			input: withNames(&categoryName, &counterpartyName),
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).Return(nil, nil)
				m.category.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.TransactionCategory) (*model.TransactionCategory, error) {
						assert.Equal(t, userID, c.UserID)
						assert.Equal(t, categoryName, c.Name)
						assert.Equal(t, enum.Expense, c.Type)
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
				expectUpdate(m)
			},
			wantResult: func(t *testing.T, result *transaction.UpdateTransactionResult) {
				require.NotNil(t, result.Transaction.TransactionCategoryID)
				assert.Equal(t, createdCategoryID, *result.Transaction.TransactionCategoryID)
				require.NotNil(t, result.Transaction.CounterpartyID)
				assert.Equal(t, createdCounterpartyID, *result.Transaction.CounterpartyID)
			},
		},
		{
			name: "category is looked up by the new transaction type",
			input: func() transaction.UpdateTransactionInput {
				i := withNames(&categoryName, nil)
				i.Type = enum.TransactionIncome
				return i
			}(),
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(&storedExpense, nil)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Income, categoryName).
					Return(&model.TransactionCategory{ID: existingCategoryID, UserID: userID, Name: categoryName, Type: enum.Income}, nil)
				expectUpdate(m)
			},
			wantResult: func(t *testing.T, result *transaction.UpdateTransactionResult) {
				assert.Equal(t, enum.TransactionIncome, result.Transaction.Type)
				require.NotNil(t, result.Transaction.TransactionCategoryID)
				assert.Equal(t, existingCategoryID, *result.Transaction.TransactionCategoryID)
			},
		},
		{
			name: "validation error - zero amount and empty account",
			input: transaction.UpdateTransactionInput{
				TransactionID: transactionID,
				UserID:        userID,
				AccountID:     uuid.Nil,
				Amount:        decimal.Zero,
				Type:          enum.TransactionIncome,
			},
			setupMocks: func(m mocks) {
				// no calls expected - Validate() must short-circuit before touching any dependency
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "accountId")
				assert.Contains(t, validationErr.Fields, "amount")
			},
		},
		{
			name:       "validation error - names too long",
			input:      withNames(&tooLongCategoryName, &tooLongCounterpartyName),
			setupMocks: func(m mocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionCategoryName")
				assert.Contains(t, validationErr.Fields, "counterpartyName")
			},
		},
		{
			name:  "account not found",
			input: baseInput,
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
			},
		},
		{
			name:  "transaction not found",
			input: withNames(&categoryName, &counterpartyName),
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(nil, nil)
				// no category/counterparty calls: nothing may be created for a missing transaction
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction.ErrTransactionNotFound)
			},
		},
		{
			name:  "transfer update not allowed",
			input: baseInput,
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID}, nil)
				m.transaction.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionIncome, TransferID: &transferID}, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, transaction.ErrTransferUpdateNotAllowed)
			},
		},
		{
			name: "negative balance not allowed",
			input: func() transaction.UpdateTransactionInput {
				i := withNames(&categoryName, &counterpartyName)
				i.IsConfirmed = true
				return i
			}(),
			setupMocks: func(m mocks) {
				m.account.EXPECT().GetByIDAndUserID(mock.Anything, accountID, userID).
					Return(&model.Account{ID: accountID, UserID: userID, IsNegativeBalanceEnabled: false}, nil)
				m.transaction.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).Return(&storedExpense, nil)
				m.transaction.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(10), nil)
				// no category/counterparty calls: nothing may be created for a rejected update
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrNegativeBalanceNotAllowed)
			},
		},
		{
			name:  "category lookup error",
			input: withNames(&categoryName, nil),
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).
					Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name:  "category create error",
			input: withNames(&categoryName, nil),
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.category.EXPECT().GetByUserIDAndTypeAndName(mock.Anything, userID, enum.Expense, categoryName).Return(nil, nil)
				m.category.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
		{
			name:  "counterparty lookup error",
			input: withNames(nil, &counterpartyName),
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).
					Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name:  "counterparty create error",
			input: withNames(nil, &counterpartyName),
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.counterparty.EXPECT().GetByUserIDAndName(mock.Anything, userID, counterpartyName).Return(nil, nil)
				m.counterparty.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
		{
			name:  "repository update error",
			input: baseInput,
			setupMocks: func(m mocks) {
				expectLoaded(m, storedExpense)
				m.transaction.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errors.New("update failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "update failed")
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

			result, err := service.UpdateTransaction(context.Background(), tt.input)

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
