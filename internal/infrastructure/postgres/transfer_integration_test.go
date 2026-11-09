//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTransferRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))

	date := testDate(2)
	created, err := repo.Create(ctx, model.NewTransfer(from.ID, to.ID, dec("100"), dec("92.5"), date))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, from.ID, fetched.FromAccountID)
	require.Equal(t, to.ID, fetched.ToAccountID)
	require.True(t, fetched.FromAmount.Equal(dec("100")))
	require.True(t, fetched.ToAmount.Equal(dec("92.5")), "cross-currency toAmount differs from fromAmount")
	require.True(t, fetched.Date.Equal(date))
}

func TestTransferRepository_Create_UnknownAccount(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	_, err := repo.Create(ctx, model.NewTransfer(from.ID, uuid.New(), dec("1"), dec("1"), testDate(0)))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "FK on to_account_id must reject unknown account, got: %v", err)
}

func TestTransferRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)

	fetched, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestTransferRepository_ExistsByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	created, err := repo.Create(ctx, model.NewTransfer(from.ID, to.ID, dec("1"), dec("1"), testDate(0)))
	require.NoError(t, err)

	exists, err := repo.ExistsByID(ctx, created.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByID(ctx, uuid.New())
	require.NoError(t, err)
	require.False(t, exists)
}

func TestTransferRepository_GetByFromAndToAccountID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	a := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	b := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	c := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	aToB, err := repo.Create(ctx, model.NewTransfer(a.ID, b.ID, dec("1"), dec("1"), testDate(0)))
	require.NoError(t, err)
	aToC, err := repo.Create(ctx, model.NewTransfer(a.ID, c.ID, dec("2"), dec("2"), testDate(0)))
	require.NoError(t, err)
	bToA, err := repo.Create(ctx, model.NewTransfer(b.ID, a.ID, dec("3"), dec("3"), testDate(0)))
	require.NoError(t, err)

	from, err := repo.GetByFromAccountID(ctx, a.ID)
	require.NoError(t, err)
	require.Len(t, from, 2)
	require.ElementsMatch(t, []uuid.UUID{aToB.ID, aToC.ID}, []uuid.UUID{from[0].ID, from[1].ID})

	to, err := repo.GetByToAccountID(ctx, a.ID)
	require.NoError(t, err)
	require.Len(t, to, 1)
	require.Equal(t, bToA.ID, to[0].ID)

	none, err := repo.GetByFromAccountID(ctx, c.ID)
	require.NoError(t, err)
	require.Empty(t, none)
}

func TestTransferRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	a := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	b := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	c := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	created, err := repo.Create(ctx, model.NewTransfer(a.ID, b.ID, dec("1"), dec("1"), testDate(5)))
	require.NoError(t, err)

	created.ToAccountID = c.ID
	created.FromAmount = dec("20")
	created.ToAmount = dec("19")
	created.Date = testDate(0)

	_, err = repo.Update(ctx, created)
	require.NoError(t, err)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, c.ID, fetched.ToAccountID)
	require.True(t, fetched.FromAmount.Equal(dec("20")))
	require.True(t, fetched.ToAmount.Equal(dec("19")))
	require.True(t, fetched.Date.Equal(testDate(0)))
}

func TestTransferRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransferRepository(testClient)
	user := createTestUser(t, ctx)
	a := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	b := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	ghost := model.NewTransfer(a.ID, b.ID, dec("1"), dec("1"), testDate(0))
	ghost.ID = uuid.New()

	_, err := repo.Update(ctx, ghost)
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err))
}

func TestTransferRepository_DeleteByID_CascadesToLinkedTransactions(t *testing.T) {
	ctx := context.Background()
	transfers := postgres.NewTransferRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	transfer, err := transfers.Create(ctx, model.NewTransfer(from.ID, to.ID, dec("40"), dec("40"), testDate(0)))
	require.NoError(t, err)

	out, err := transactions.Create(ctx, model.NewTransferOutTransaction(from.ID, dec("40"), "out", testDate(0), true, transfer.ID))
	require.NoError(t, err)
	in, err := transactions.Create(ctx, model.NewTransferInTransaction(to.ID, dec("40"), "in", testDate(0), true, transfer.ID))
	require.NoError(t, err)
	plain := createTestTransaction(t, ctx, from.ID, enum.TransactionExpense, dec("1"), testDate(0), true, nil)

	require.NoError(t, transfers.DeleteByID(ctx, transfer.ID))

	gone, err := transfers.GetByID(ctx, transfer.ID)
	require.NoError(t, err)
	require.Nil(t, gone)

	for _, id := range []uuid.UUID{out.ID, in.ID} {
		tx, err := transactions.GetByID(ctx, id)
		require.NoError(t, err)
		require.Nil(t, tx, "transfer transactions are removed with the transfer")
	}

	kept, err := transactions.GetByID(ctx, plain.ID)
	require.NoError(t, err)
	require.NotNil(t, kept)

	balance, err := transactions.GetBalanceByAccountID(ctx, from.ID)
	require.NoError(t, err)
	require.True(t, balance.Equal(dec("99")), "balance is restored after transfer delete, got %s", balance)
}

func TestTransferRepository_TransferTransactionsAffectBalances(t *testing.T) {
	ctx := context.Background()
	transfers := postgres.NewTransferRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))

	transfer, err := transfers.Create(ctx, model.NewTransfer(from.ID, to.ID, dec("40"), dec("37"), testDate(0)))
	require.NoError(t, err)
	_, err = transactions.Create(ctx, model.NewTransferOutTransaction(from.ID, dec("40"), "out", testDate(0), true, transfer.ID))
	require.NoError(t, err)
	_, err = transactions.Create(ctx, model.NewTransferInTransaction(to.ID, dec("37"), "in", testDate(0), true, transfer.ID))
	require.NoError(t, err)

	fromBalance, err := transactions.GetBalanceByAccountID(ctx, from.ID)
	require.NoError(t, err)
	require.True(t, fromBalance.Equal(dec("60")), "got %s", fromBalance)

	toBalance, err := transactions.GetBalanceByAccountID(ctx, to.ID)
	require.NoError(t, err)
	require.True(t, toBalance.Equal(dec("37")), "got %s", toBalance)
}
