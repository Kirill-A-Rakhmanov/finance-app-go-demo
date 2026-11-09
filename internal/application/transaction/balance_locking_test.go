package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type lockingScenario struct {
	accountID     uuid.UUID
	userID        uuid.UUID
	transactionID uuid.UUID
	accounts      *repomock.MockAccountRepository
	transactions  *repomock.MockTransactionRepository
	txManager     *transactionmock.MockManager
	locked        bool
}

func newLockingScenario(t *testing.T) *lockingScenario {
	t.Helper()

	s := &lockingScenario{
		accountID:     uuid.New(),
		userID:        uuid.New(),
		transactionID: uuid.New(),
		accounts:      repomock.NewMockAccountRepository(t),
		transactions:  repomock.NewMockTransactionRepository(t),
		txManager:     transactionmock.NewMockManager(t),
	}

	s.txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		})

	return s
}

func (s *lockingScenario) service() *transaction.Service {
	return transaction.NewService(s.transactions, s.accounts, nil, nil, s.txManager, nil)
}

func (s *lockingScenario) expectAccount(allowsNegative bool) {
	s.accounts.EXPECT().GetByIDAndUserID(mock.Anything, s.accountID, s.userID).
		Return(&model.Account{ID: s.accountID, UserID: s.userID, IsNegativeBalanceEnabled: allowsNegative}, nil)
}

func (s *lockingScenario) expectLock(result error) {
	s.txManager.EXPECT().Lock(mock.Anything, model.AccountBalanceLockKey(s.accountID)).
		RunAndReturn(func(ctx context.Context, key string) error {
			s.locked = true
			return result
		})
}

func (s *lockingScenario) expectBalance(t *testing.T, balance int64) {
	s.transactions.EXPECT().GetBalanceByAccountID(mock.Anything, s.accountID).
		RunAndReturn(func(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
			assert.True(t, s.locked, "the balance must be read only after the account is locked")
			return decimal.NewFromInt(balance), nil
		})
}

func TestCreateExpenseTransaction_LocksTheAccountBeforeReadingTheBalance(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccount(false)
	s.expectLock(nil)
	s.expectBalance(t, 100)
	s.transactions.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) { return tx, nil })

	_, err := s.service().CreateExpenseTransaction(context.Background(), transaction.CreateExpenseTransactionInput{
		UserID: s.userID, AccountID: s.accountID, Amount: decimal.NewFromInt(40), Date: time.Now(),
	})

	require.NoError(t, err)
	assert.True(t, s.locked)
}

func TestCreateExpenseTransaction_AccountThatAllowsNegativeBalanceIsNotLocked(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccount(true)
	s.transactions.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) { return tx, nil })

	_, err := s.service().CreateExpenseTransaction(context.Background(), transaction.CreateExpenseTransactionInput{
		UserID: s.userID, AccountID: s.accountID, Amount: decimal.NewFromInt(40), Date: time.Now(),
	})

	require.NoError(t, err)
	assert.False(t, s.locked, "no lock when there is no balance rule to protect")
}

func TestCreateExpenseTransaction_LockFailureStopsTheOperation(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccount(false)
	s.expectLock(errors.New("lock timeout"))

	result, err := s.service().CreateExpenseTransaction(context.Background(), transaction.CreateExpenseTransactionInput{
		UserID: s.userID, AccountID: s.accountID, Amount: decimal.NewFromInt(40), Date: time.Now(),
	})

	assert.EqualError(t, err, "lock timeout")
	assert.Nil(t, result)
}

func TestUpdateTransaction_LocksTheAccountBeforeReadingTheBalance(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccount(false)
	s.transactions.EXPECT().GetByIDAndUserID(mock.Anything, s.transactionID, s.userID).
		Return(&model.Transaction{ID: s.transactionID, AccountID: s.accountID, Type: enum.TransactionExpense, Amount: decimal.NewFromInt(10), IsConfirmed: true}, nil)
	s.expectLock(nil)
	s.expectBalance(t, 100)
	s.transactions.EXPECT().Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) { return tx, nil })

	_, err := s.service().UpdateTransaction(context.Background(), transaction.UpdateTransactionInput{
		TransactionID: s.transactionID, UserID: s.userID, AccountID: s.accountID,
		Type: enum.TransactionExpense, Amount: decimal.NewFromInt(40), Date: time.Now(), IsConfirmed: true,
	})

	require.NoError(t, err)
	assert.True(t, s.locked)
}

func TestUpdateTransaction_IncomeIsNotLocked(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccount(false)
	s.transactions.EXPECT().GetByIDAndUserID(mock.Anything, s.transactionID, s.userID).
		Return(&model.Transaction{ID: s.transactionID, AccountID: s.accountID, Type: enum.TransactionIncome}, nil)
	s.transactions.EXPECT().Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) { return tx, nil })

	_, err := s.service().UpdateTransaction(context.Background(), transaction.UpdateTransactionInput{
		TransactionID: s.transactionID, UserID: s.userID, AccountID: s.accountID,
		Type: enum.TransactionIncome, Amount: decimal.NewFromInt(40), Date: time.Now(),
	})

	require.NoError(t, err)
	assert.False(t, s.locked)
}

func TestUpdateTransaction_LockFailureStopsTheOperation(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccount(false)
	s.transactions.EXPECT().GetByIDAndUserID(mock.Anything, s.transactionID, s.userID).
		Return(&model.Transaction{ID: s.transactionID, AccountID: s.accountID, Type: enum.TransactionExpense, Amount: decimal.NewFromInt(10), IsConfirmed: true}, nil)
	s.expectLock(errors.New("lock timeout"))

	_, err := s.service().UpdateTransaction(context.Background(), transaction.UpdateTransactionInput{
		TransactionID: s.transactionID, UserID: s.userID, AccountID: s.accountID,
		Type: enum.TransactionExpense, Amount: decimal.NewFromInt(40), Date: time.Now(), IsConfirmed: true,
	})

	assert.EqualError(t, err, "lock timeout")
}

func TestDeleteTransaction_RunsInOneTransactionAndLocksBeforeReadingTheBalance(t *testing.T) {
	s := newLockingScenario(t)
	s.transactions.EXPECT().GetByIDAndUserID(mock.Anything, s.transactionID, s.userID).
		Return(&model.Transaction{ID: s.transactionID, AccountID: s.accountID, Type: enum.TransactionIncome, Amount: decimal.NewFromInt(30), IsConfirmed: true}, nil)
	s.expectAccount(false)
	s.expectLock(nil)
	s.expectBalance(t, 100)
	s.transactions.EXPECT().DeleteByID(mock.Anything, s.transactionID).Return(nil)

	result, err := s.service().DeleteTransaction(context.Background(), transaction.DeleteTransactionInput{
		TransactionID: s.transactionID, UserID: s.userID,
	})

	require.NoError(t, err)
	assert.True(t, decimal.NewFromInt(70).Equal(result.Balance))
	assert.True(t, s.locked)
}

func TestDeleteTransaction_AccountThatAllowsNegativeBalanceIsNotLocked(t *testing.T) {
	s := newLockingScenario(t)
	s.transactions.EXPECT().GetByIDAndUserID(mock.Anything, s.transactionID, s.userID).
		Return(&model.Transaction{ID: s.transactionID, AccountID: s.accountID, Type: enum.TransactionIncome, Amount: decimal.NewFromInt(30), IsConfirmed: true}, nil)
	s.expectAccount(true)
	s.transactions.EXPECT().GetBalanceByAccountID(mock.Anything, s.accountID).Return(decimal.NewFromInt(10), nil)
	s.transactions.EXPECT().DeleteByID(mock.Anything, s.transactionID).Return(nil)

	_, err := s.service().DeleteTransaction(context.Background(), transaction.DeleteTransactionInput{
		TransactionID: s.transactionID, UserID: s.userID,
	})

	require.NoError(t, err)
	assert.False(t, s.locked)
}

func TestDeleteTransaction_LockFailureStopsTheOperation(t *testing.T) {
	s := newLockingScenario(t)
	s.transactions.EXPECT().GetByIDAndUserID(mock.Anything, s.transactionID, s.userID).
		Return(&model.Transaction{ID: s.transactionID, AccountID: s.accountID, Type: enum.TransactionIncome, Amount: decimal.NewFromInt(30), IsConfirmed: true}, nil)
	s.expectAccount(false)
	s.expectLock(errors.New("lock timeout"))

	result, err := s.service().DeleteTransaction(context.Background(), transaction.DeleteTransactionInput{
		TransactionID: s.transactionID, UserID: s.userID,
	})

	assert.EqualError(t, err, "lock timeout")
	assert.Nil(t, result)
}
