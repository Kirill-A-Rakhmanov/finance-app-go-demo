package transaction_test

import (
	"context"
	"slices"
	"testing"
	"time"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
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

func TestService_UpdateTransaction_BalanceRules(t *testing.T) {
	userID := uuid.New()
	transactionID := uuid.New()
	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)
	eur, err := valueObject.NewCurrency(valueObject.CurrencyEUR)
	require.NoError(t, err)

	type stored struct {
		account     string
		type_       enum.TransactionType
		amount      int64
		isConfirmed bool
	}
	type update struct {
		account     string
		type_       enum.TransactionType
		amount      int64
		isConfirmed bool
	}
	type account struct {
		allowsNegative bool
		currency       valueObject.Currency
	}

	expense, income := enum.TransactionExpense, enum.TransactionIncome
	strict := map[string]account{"a": {currency: usd}, "b": {currency: usd}}

	tests := []struct {
		name     string
		accounts map[string]account
		stored   stored
		update   update
		balances map[string]int64
		locked   []string
		wantErr  error
	}{
		{
			name:     "lowering a confirmed expense needs no check even with nothing left on the account",
			accounts: strict,
			stored:   stored{"a", expense, 100, true},
			update:   update{"a", expense, 90, true},
		},
		{
			name:     "raising an expense is checked against the balance that already includes the old amount",
			accounts: strict,
			stored:   stored{"a", expense, 100, true},
			update:   update{"a", expense, 130, true},
			balances: map[string]int64{"a": 30},
			locked:   []string{"a"},
		},
		{
			name:     "raising an expense by more than the balance is refused",
			accounts: strict,
			stored:   stored{"a", expense, 100, true},
			update:   update{"a", expense, 150, true},
			balances: map[string]int64{"a": 30},
			locked:   []string{"a"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "lowering an income below what was already spent is refused",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"a", income, 60, true},
			balances: map[string]int64{"a": 30},
			locked:   []string{"a"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "lowering an income that still leaves enough is allowed",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"a", income, 60, true},
			balances: map[string]int64{"a": 100},
			locked:   []string{"a"},
		},
		{
			name:     "raising an income needs no check",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"a", income, 500, true},
		},
		{
			name:     "turning an income into an expense removes it and spends the same amount",
			accounts: strict,
			stored:   stored{"a", income, 50, true},
			update:   update{"a", expense, 50, true},
			balances: map[string]int64{"a": 80},
			locked:   []string{"a"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "turning an income into an expense is allowed when the balance covers both",
			accounts: strict,
			stored:   stored{"a", income, 50, true},
			update:   update{"a", expense, 50, true},
			balances: map[string]int64{"a": 100},
			locked:   []string{"a"},
		},
		{
			name:     "turning an expense into an income needs no check",
			accounts: strict,
			stored:   stored{"a", expense, 50, true},
			update:   update{"a", income, 50, true},
		},
		{
			name:     "unconfirming an income takes its money away",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"a", income, 100, false},
			balances: map[string]int64{"a": 50},
			locked:   []string{"a"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "confirming an expense spends the money",
			accounts: strict,
			stored:   stored{"a", expense, 40, false},
			update:   update{"a", expense, 40, true},
			balances: map[string]int64{"a": 30},
			locked:   []string{"a"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "an unconfirmed transaction stays out of the balance, however big it is",
			accounts: strict,
			stored:   stored{"a", expense, 10, false},
			update:   update{"a", expense, 1000, false},
		},
		{
			name:     "changing only other fields never reads the balance, even on an overdrawn account",
			accounts: strict,
			stored:   stored{"a", expense, 40, true},
			update:   update{"a", expense, 40, true},
		},
		{
			name:     "an account that allows a negative balance is neither checked nor locked",
			accounts: map[string]account{"a": {allowsNegative: true, currency: usd}},
			stored:   stored{"a", expense, 10, true},
			update:   update{"a", expense, 1000, true},
		},
		{
			name:     "moving an expense checks only the account that receives it",
			accounts: strict,
			stored:   stored{"a", expense, 100, true},
			update:   update{"b", expense, 100, true},
			balances: map[string]int64{"b": 150},
			locked:   []string{"b"},
		},
		{
			name:     "moving an expense to an account that cannot afford it is refused",
			accounts: strict,
			stored:   stored{"a", expense, 100, true},
			update:   update{"b", expense, 100, true},
			balances: map[string]int64{"b": 50},
			locked:   []string{"b"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "moving an income checks only the account it leaves",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"b", income, 100, true},
			balances: map[string]int64{"a": 100},
			locked:   []string{"a"},
		},
		{
			name:     "moving an income that was already spent is refused",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"b", income, 100, true},
			balances: map[string]int64{"a": 60},
			locked:   []string{"a"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "moving an income into an expense locks and checks both accounts",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"b", expense, 20, true},
			balances: map[string]int64{"a": 100, "b": 20},
			locked:   []string{"a", "b"},
		},
		{
			name:     "the account the transaction leaves is protected too",
			accounts: strict,
			stored:   stored{"a", income, 100, true},
			update:   update{"b", expense, 20, true},
			balances: map[string]int64{"a": 40, "b": 500},
			locked:   []string{"a", "b"},
			wantErr:  apperror.ErrNegativeBalanceNotAllowed,
		},
		{
			name:     "moving to an account in another currency is refused before anything is locked",
			accounts: map[string]account{"a": {currency: usd}, "b": {currency: eur}},
			stored:   stored{"a", expense, 10, true},
			update:   update{"b", expense, 10, true},
			wantErr:  transaction.ErrMoveToOtherCurrencyNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := map[string]uuid.UUID{"a": uuid.New(), "b": uuid.New()}
			accounts := repomock.NewMockAccountRepository(t)
			transactions := repomock.NewMockTransactionRepository(t)
			txManager := transactionmock.NewMockManager(t)

			txManager.EXPECT().
				Wrap(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
					return txFunc(ctx)
				})

			for label, acc := range tt.accounts {
				if label != tt.update.account && label != tt.stored.account {
					continue
				}
				accounts.EXPECT().GetByIDAndUserID(mock.Anything, ids[label], userID).
					Return(&model.Account{ID: ids[label], UserID: userID, Currency: acc.currency, IsNegativeBalanceEnabled: acc.allowsNegative}, nil)
			}

			transactions.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
				Return(&model.Transaction{
					ID: transactionID, AccountID: ids[tt.stored.account], Type: tt.stored.type_,
					Amount: decimal.NewFromInt(tt.stored.amount), IsConfirmed: tt.stored.isConfirmed,
				}, nil)

			var lockedKeys []string
			locksTaken := map[string]bool{}
			for _, label := range tt.locked {
				txManager.EXPECT().Lock(mock.Anything, model.AccountBalanceLockKey(ids[label])).
					RunAndReturn(func(ctx context.Context, key string) error {
						lockedKeys = append(lockedKeys, key)
						locksTaken[label] = true
						return nil
					})
			}

			for label, balance := range tt.balances {
				balanceCall := transactions.EXPECT().GetBalanceByAccountID(mock.Anything, ids[label]).
					RunAndReturn(func(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
						assert.True(t, locksTaken[label], "the balance of %q must be read under its lock", label)
						return decimal.NewFromInt(balance), nil
					})
				if tt.wantErr != nil {
					// the accounts are checked in lock-key order, so a refusal may come before the other one is read
					balanceCall.Maybe()
				}
			}

			if tt.wantErr == nil {
				transactions.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
						return tx, nil
					})
			}

			service := transaction.NewService(transactions, accounts, nil, nil, txManager, nil)

			result, err := service.UpdateTransaction(context.Background(), transaction.UpdateTransactionInput{
				TransactionID: transactionID, UserID: userID, AccountID: ids[tt.update.account],
				Type: tt.update.type_, Amount: decimal.NewFromInt(tt.update.amount),
				IsConfirmed: tt.update.isConfirmed, Date: time.Now(),
			})

			assert.True(t, slices.IsSorted(lockedKeys), "several locks are always taken in one order: %v", lockedKeys)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, ids[tt.update.account], result.Transaction.AccountID, "the transaction ends up in the requested account")
		})
	}
}

func TestService_UpdateTransaction_SourceAccountMustStillExistToMove(t *testing.T) {
	userID := uuid.New()
	transactionID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()

	accounts := repomock.NewMockAccountRepository(t)
	transactions := repomock.NewMockTransactionRepository(t)
	txManager := transactionmock.NewMockManager(t)
	txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error { return txFunc(ctx) })

	accounts.EXPECT().GetByIDAndUserID(mock.Anything, targetID, userID).
		Return(&model.Account{ID: targetID, UserID: userID}, nil)
	accounts.EXPECT().GetByIDAndUserID(mock.Anything, sourceID, userID).Return(nil, nil)
	transactions.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, userID).
		Return(&model.Transaction{ID: transactionID, AccountID: sourceID, Type: enum.TransactionExpense}, nil)

	service := transaction.NewService(transactions, accounts, nil, nil, txManager, nil)

	_, err := service.UpdateTransaction(context.Background(), transaction.UpdateTransactionInput{
		TransactionID: transactionID, UserID: userID, AccountID: targetID,
		Type: enum.TransactionExpense, Amount: decimal.NewFromInt(1), Date: time.Now(),
	})

	assert.ErrorIs(t, err, apperror.ErrAccountNotFound)
}
