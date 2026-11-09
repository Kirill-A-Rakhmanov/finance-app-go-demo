//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	pkgtx "finance-app-go-demo/pkg/transaction"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTransactionManager_Wrap_CommitsOnSuccess(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	accounts := postgres.NewAccountRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	newTx := model.NewIncomeTransaction(account.ID, nil, nil, dec("5"), nil, testDate(0), true)

	err := manager.Wrap(ctx, func(ctx context.Context) error {
		created, err := transactions.Create(ctx, newTx)
		if err != nil {
			return err
		}
		newTx.ID = created.ID

		return nil
	})
	require.NoError(t, err)

	fetched, err := transactions.GetByID(ctx, newTx.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched, "committed write must be visible outside the transaction")

	exists, err := accounts.ExistsByID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestTransactionManager_Wrap_RollsBackOnError(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	transfers := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	boom := errors.New("boom")

	err := manager.Wrap(ctx, func(ctx context.Context) error {
		transfer, err := transfers.Create(ctx, model.NewTransfer(from.ID, to.ID, dec("10"), dec("10"), testDate(0)))
		if err != nil {
			return err
		}
		if _, err := transactions.Create(ctx, model.NewTransferOutTransaction(from.ID, dec("10"), "out", testDate(0), true, transfer.ID)); err != nil {
			return err
		}

		return boom
	})
	require.ErrorIs(t, err, boom, "the original error must be returned")

	fromTransfers, err := transfers.GetByFromAccountID(ctx, from.ID)
	require.NoError(t, err)
	require.Empty(t, fromTransfers, "transfer must be rolled back")

	txs, err := transactions.GetByAccountID(ctx, from.ID)
	require.NoError(t, err)
	require.Empty(t, txs, "transactions must be rolled back")
}

func TestTransactionManager_Wrap_ReadsOwnUncommittedWrites(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	err := manager.Wrap(ctx, func(txCtx context.Context) error {
		if _, err := transactions.Create(txCtx, model.NewIncomeTransaction(account.ID, nil, nil, dec("25"), nil, testDate(0), true)); err != nil {
			return err
		}

		inside, err := transactions.GetBalanceByAccountID(txCtx, account.ID)
		require.NoError(t, err)
		require.True(t, inside.Equal(dec("25")), "balance inside tx sees its own write, got %s", inside)

		outside, err := transactions.GetBalanceByAccountID(ctx, account.ID)
		require.NoError(t, err)
		require.True(t, outside.IsZero(), "uncommitted write must be invisible outside, got %s", outside)

		return nil
	})
	require.NoError(t, err)

	after, err := transactions.GetBalanceByAccountID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, after.Equal(dec("25")), "got %s", after)
}

func lockKey() string {
	return "test-lock:" + uuid.NewString()
}

func TestTransactionManager_Lock_RequiresATransaction(t *testing.T) {
	manager := pkgtx.NewTransactionManager(testClient)

	err := manager.Lock(context.Background(), lockKey())

	require.ErrorIs(t, err, pkgtx.ErrNoActiveTransaction)
}

func TestTransactionManager_Lock_IsReentrantWithinATransaction(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	key := lockKey()

	err := manager.Wrap(ctx, func(ctx context.Context) error {
		if err := manager.Lock(ctx, key); err != nil {
			return err
		}

		return manager.Lock(ctx, key)
	})

	require.NoError(t, err, "a transaction can take the same lock again without waiting for itself")
}

func TestTransactionManager_Lock_BlocksAnotherTransactionUntilCommit(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	key := lockKey()

	firstLocked := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondLocked := make(chan struct{})
	errs := make(chan error, 2)

	go func() {
		errs <- manager.Wrap(ctx, func(ctx context.Context) error {
			if err := manager.Lock(ctx, key); err != nil {
				return err
			}
			close(firstLocked)
			<-releaseFirst

			return nil
		})
	}()

	<-firstLocked

	go func() {
		errs <- manager.Wrap(ctx, func(ctx context.Context) error {
			if err := manager.Lock(ctx, key); err != nil {
				return err
			}
			close(secondLocked)

			return nil
		})
	}()

	select {
	case <-secondLocked:
		t.Fatal("the second transaction got the lock while the first one still held it")
	case <-time.After(500 * time.Millisecond):
	}

	close(releaseFirst)

	select {
	case <-secondLocked:
	case <-time.After(5 * time.Second):
		t.Fatal("the second transaction never got the lock after the first one committed")
	}

	require.NoError(t, <-errs)
	require.NoError(t, <-errs)
}

func TestTransactionManager_Lock_ReleasesOnRollback(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	key := lockKey()
	boom := errors.New("boom")

	err := manager.Wrap(ctx, func(ctx context.Context) error {
		if err := manager.Lock(ctx, key); err != nil {
			return err
		}

		return boom
	})
	require.ErrorIs(t, err, boom)

	acquired := make(chan error, 1)
	go func() {
		acquired <- manager.Wrap(ctx, func(ctx context.Context) error {
			return manager.Lock(ctx, key)
		})
	}()

	select {
	case err := <-acquired:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the lock was not released by the rolled back transaction")
	}
}

func TestTransactionManager_Lock_DifferentKeysDoNotBlockEachOther(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)

	firstLocked := make(chan struct{})
	releaseFirst := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		done <- manager.Wrap(ctx, func(ctx context.Context) error {
			if err := manager.Lock(ctx, lockKey()); err != nil {
				return err
			}
			close(firstLocked)
			<-releaseFirst

			return nil
		})
	}()

	<-firstLocked

	otherDone := make(chan error, 1)
	go func() {
		otherDone <- manager.Wrap(ctx, func(ctx context.Context) error {
			return manager.Lock(ctx, lockKey())
		})
	}()

	select {
	case err := <-otherDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("a lock on another key was blocked")
	}

	close(releaseFirst)
	require.NoError(t, <-done)
}

func TestTransactionManager_Lock_ConcurrentExpensesCannotOverdrawAnAccount(t *testing.T) {
	ctx := context.Background()
	manager := pkgtx.NewTransactionManager(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))

	errInsufficient := errors.New("insufficient funds")
	spend := func() error {
		return manager.Wrap(ctx, func(ctx context.Context) error {
			if err := manager.Lock(ctx, model.AccountBalanceLockKey(account.ID)); err != nil {
				return err
			}

			balance, err := transactions.GetBalanceByAccountID(ctx, account.ID)
			if err != nil {
				return err
			}
			if balance.Sub(dec("60")).IsNegative() {
				return errInsufficient
			}

			time.Sleep(50 * time.Millisecond)

			_, err = transactions.Create(ctx, model.NewExpenseTransaction(account.ID, nil, nil, dec("60"), nil, testDate(0), true))

			return err
		})
	}

	const attempts = 8
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- spend()
		}()
	}
	wg.Wait()
	close(results)

	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
			continue
		}
		require.ErrorIs(t, err, errInsufficient)
	}

	require.Equal(t, 1, succeeded, "only one of the %d concurrent 60-unit expenses fits into a balance of 100", attempts)

	balance, err := transactions.GetBalanceByAccountID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, balance.Equal(dec("40")), "got %s", balance)
}
