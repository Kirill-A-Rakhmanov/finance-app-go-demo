//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestAccountRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)

	eur, err := valueObject.NewCurrency(valueObject.CurrencyEUR)
	require.NoError(t, err)

	account := model.NewAccount(user.ID, "Wallet", "wallet-icon", eur, false, false, false, decimal.NewFromInt(100))

	created, err := repo.Create(ctx, account)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)
	require.Equal(t, valueObject.CurrencyEUR, created.Currency.Code)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, "Wallet", fetched.Name)
	require.True(t, created.Adjustment.Equal(fetched.Adjustment))
}

func TestAccountRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)

	fetched, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestAccountRepository_GetByIDAndUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)
	otherUser := createTestUser(t, ctx)

	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	created, err := repo.Create(ctx, model.NewAccount(user.ID, "Wallet", "icon", usd, false, false, false, decimal.Zero))
	require.NoError(t, err)

	found, err := repo.GetByIDAndUserID(ctx, created.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, created.ID, found.ID)

	notFound, err := repo.GetByIDAndUserID(ctx, created.ID, otherUser.ID)
	require.NoError(t, err)
	require.Nil(t, notFound)
}

func TestAccountRepository_Update_PreservesCurrency(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)

	eur, err := valueObject.NewCurrency(valueObject.CurrencyEUR)
	require.NoError(t, err)

	created, err := repo.Create(ctx, model.NewAccount(user.ID, "Wallet", "icon", eur, false, false, false, decimal.Zero))
	require.NoError(t, err)
	require.Equal(t, valueObject.CurrencyEUR, created.Currency.Code)

	created.Name = "Renamed Wallet"

	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, "Renamed Wallet", updated.Name)
	require.Equal(t, valueObject.CurrencyEUR, updated.Currency.Code, "Update must not silently change the account currency")

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, valueObject.CurrencyEUR, fetched.Currency.Code)
}

func TestAccountRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)

	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	created, err := repo.Create(ctx, model.NewAccount(user.ID, "ToDelete", "icon", usd, false, false, false, decimal.Zero))
	require.NoError(t, err)

	require.NoError(t, repo.DeleteByID(ctx, created.ID))

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestAccountRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)

	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.NewAccount(user.ID, "Acc1", "icon", usd, false, false, false, decimal.Zero))
	require.NoError(t, err)
	_, err = repo.Create(ctx, model.NewAccount(user.ID, "Acc2", "icon", usd, false, false, false, decimal.Zero))
	require.NoError(t, err)

	accounts, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, accounts, 2)
}

func TestAccountRepository_ExistsByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)

	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	created, err := repo.Create(ctx, model.NewAccount(user.ID, "Wallet", "icon", usd, false, false, false, decimal.Zero))
	require.NoError(t, err)

	exists, err := repo.ExistsByID(ctx, created.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByID(ctx, uuid.New())
	require.NoError(t, err)
	require.False(t, exists)
}

func TestAccountRepository_GetByIDsAndUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	first := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, decimal.Zero)
	second := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, decimal.Zero)
	createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, decimal.Zero)
	foreign := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, decimal.Zero)

	t.Run("returns only the requested accounts", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{first.ID, second.ID}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 2)
		require.ElementsMatch(t, []uuid.UUID{first.ID, second.ID}, []uuid.UUID{found[0].ID, found[1].ID})
	})

	t.Run("another user's account is not returned", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{first.ID, foreign.ID}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 1)
		require.Equal(t, first.ID, found[0].ID)
	})

	t.Run("unknown ids are skipped", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{first.ID, uuid.New()}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 1)
	})

	t.Run("duplicate ids return the account once", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{first.ID, first.ID}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 1)
	})

	t.Run("currency is mapped", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{second.ID}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 1)
		require.Equal(t, valueObject.CurrencyEUR, found[0].Currency.Code)
	})

	t.Run("empty id list returns nothing", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, nil, user.ID)
		require.NoError(t, err)
		require.Empty(t, found)
	})
}
