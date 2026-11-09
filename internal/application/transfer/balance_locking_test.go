package transfer_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

type lockingScenario struct {
	userID       uuid.UUID
	fromID       uuid.UUID
	toID         uuid.UUID
	transferID   uuid.UUID
	accounts     *repomock.MockAccountRepository
	transactions *repomock.MockTransactionRepository
	transfers    *repomock.MockTransferRepository
	txManager    *transactionmock.MockManager
	lockedKeys   []string
}

func newLockingScenario(t *testing.T) *lockingScenario {
	t.Helper()

	s := &lockingScenario{
		userID:       uuid.New(),
		fromID:       uuid.New(),
		toID:         uuid.New(),
		transferID:   uuid.New(),
		accounts:     repomock.NewMockAccountRepository(t),
		transactions: repomock.NewMockTransactionRepository(t),
		transfers:    repomock.NewMockTransferRepository(t),
		txManager:    transactionmock.NewMockManager(t),
	}

	s.txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		})

	return s
}

func (s *lockingScenario) service() *transfer.Service {
	return transfer.NewService(s.transfers, s.transactions, s.accounts, s.txManager)
}

func (s *lockingScenario) expectLock(accountID uuid.UUID, result error) {
	s.txManager.EXPECT().Lock(mock.Anything, model.AccountBalanceLockKey(accountID)).
		RunAndReturn(func(ctx context.Context, key string) error {
			s.lockedKeys = append(s.lockedKeys, key)
			return result
		})
}

func (s *lockingScenario) expectBalance(t *testing.T, accountID uuid.UUID, balance int64) {
	s.transactions.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).
		RunAndReturn(func(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
			assert.Contains(t, s.lockedKeys, model.AccountBalanceLockKey(accountID), "the balance must be read only after the account is locked")
			return decimal.NewFromInt(balance), nil
		})
}

func (s *lockingScenario) expectAccounts(fromAllowsNegative bool) {
	usd, _ := valueObject.NewCurrency(valueObject.CurrencyUSD)

	s.accounts.EXPECT().GetByIDAndUserID(mock.Anything, s.fromID, s.userID).
		Return(&model.Account{ID: s.fromID, UserID: s.userID, Currency: usd, IsNegativeBalanceEnabled: fromAllowsNegative}, nil)
	s.accounts.EXPECT().GetByIDAndUserID(mock.Anything, s.toID, s.userID).
		Return(&model.Account{ID: s.toID, UserID: s.userID, Currency: usd}, nil)
}

func (s *lockingScenario) createInput() transfer.CreateTransferInput {
	return transfer.CreateTransferInput{
		UserID: s.userID, FromAccountID: s.fromID, ToAccountID: s.toID,
		FromAmount: decimal.NewFromInt(20), ToAmount: decimal.NewFromInt(20), Date: time.Now(),
	}
}

func (s *lockingScenario) expectCreateWrites() {
	s.transfers.EXPECT().Create(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tr *model.Transfer) (*model.Transfer, error) {
			tr.ID = s.transferID
			return tr, nil
		})
	s.transactions.EXPECT().BulkCreate(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, ts []model.Transaction) ([]model.Transaction, error) { return ts, nil })
}

func TestCreateTransfer_LocksOnlyTheSourceAccountBeforeReadingItsBalance(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccounts(false)
	s.expectLock(s.fromID, nil)
	s.expectBalance(t, s.fromID, 100)
	s.expectCreateWrites()

	_, err := s.service().CreateTransfer(context.Background(), s.createInput())

	require.NoError(t, err)
	assert.Equal(t, []string{model.AccountBalanceLockKey(s.fromID)}, s.lockedKeys, "the destination only gains money, so it needs no lock")
}

func TestCreateTransfer_SourceThatAllowsNegativeBalanceIsNotLocked(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccounts(true)
	s.expectCreateWrites()

	_, err := s.service().CreateTransfer(context.Background(), s.createInput())

	require.NoError(t, err)
	assert.Empty(t, s.lockedKeys)
}

func TestCreateTransfer_LockFailureStopsTheOperation(t *testing.T) {
	s := newLockingScenario(t)
	s.expectAccounts(false)
	s.expectLock(s.fromID, errors.New("lock timeout"))

	result, err := s.service().CreateTransfer(context.Background(), s.createInput())

	assert.EqualError(t, err, "lock timeout")
	assert.Nil(t, result)
}

func (s *lockingScenario) expectStoredTransfer(toAllowsNegative bool) {
	s.transfers.EXPECT().GetByID(mock.Anything, s.transferID).
		Return(&model.Transfer{ID: s.transferID, FromAccountID: s.fromID, ToAccountID: s.toID, ToAmount: decimal.NewFromInt(20)}, nil)
	s.accounts.EXPECT().GetByIDAndUserID(mock.Anything, s.toID, s.userID).
		Return(&model.Account{ID: s.toID, UserID: s.userID, IsNegativeBalanceEnabled: toAllowsNegative}, nil)
}

func TestDeleteTransfer_RunsInOneTransactionAndLocksTheDestinationBeforeReadingItsBalance(t *testing.T) {
	s := newLockingScenario(t)
	s.expectStoredTransfer(false)
	s.expectLock(s.toID, nil)
	s.expectBalance(t, s.toID, 100)
	s.transfers.EXPECT().DeleteByID(mock.Anything, s.transferID).Return(nil)

	_, err := s.service().DeleteTransfer(context.Background(), transfer.DeleteTransferInput{UserID: s.userID, TransferID: s.transferID})

	require.NoError(t, err)
	assert.Equal(t, []string{model.AccountBalanceLockKey(s.toID)}, s.lockedKeys)
}

func TestDeleteTransfer_DestinationThatAllowsNegativeBalanceIsNotLocked(t *testing.T) {
	s := newLockingScenario(t)
	s.expectStoredTransfer(true)
	s.transfers.EXPECT().DeleteByID(mock.Anything, s.transferID).Return(nil)

	_, err := s.service().DeleteTransfer(context.Background(), transfer.DeleteTransferInput{UserID: s.userID, TransferID: s.transferID})

	require.NoError(t, err)
	assert.Empty(t, s.lockedKeys)
}

func TestDeleteTransfer_LockFailureStopsTheOperation(t *testing.T) {
	s := newLockingScenario(t)
	s.expectStoredTransfer(false)
	s.expectLock(s.toID, errors.New("lock timeout"))

	result, err := s.service().DeleteTransfer(context.Background(), transfer.DeleteTransferInput{UserID: s.userID, TransferID: s.transferID})

	assert.EqualError(t, err, "lock timeout")
	assert.Nil(t, result)
}
