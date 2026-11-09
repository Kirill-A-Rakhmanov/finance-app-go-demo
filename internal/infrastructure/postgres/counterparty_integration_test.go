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

func TestCounterpartyRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	user := createTestUser(t, ctx)

	created, err := repo.Create(ctx, model.NewCounterparty(user.ID, "Cafe"))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, user.ID, fetched.UserID)
	require.Equal(t, "Cafe", fetched.Name)
}

func TestCounterpartyRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)

	fetched, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestCounterpartyRepository_GetByIDAndUserID_OtherUser(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	counterparty := createTestCounterparty(t, ctx, user.ID)

	own, err := repo.GetByIDAndUserID(ctx, counterparty.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, own)

	foreign, err := repo.GetByIDAndUserID(ctx, counterparty.ID, other.ID)
	require.NoError(t, err)
	require.Nil(t, foreign, "another user's counterparty must not be reachable")
}

func TestCounterpartyRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	first := createTestCounterparty(t, ctx, user.ID)
	second := createTestCounterparty(t, ctx, user.ID)
	createTestCounterparty(t, ctx, other.ID)

	result, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, result, 2)

	ids := []uuid.UUID{result[0].ID, result[1].ID}
	require.ElementsMatch(t, []uuid.UUID{first.ID, second.ID}, ids)
}

func TestCounterpartyRepository_Create_DuplicateNamePerUser(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	_, err := repo.Create(ctx, model.NewCounterparty(user.ID, "Cafe"))
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.NewCounterparty(user.ID, "Cafe"))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "(user_id, name) must be unique, got: %v", err)

	_, err = repo.Create(ctx, model.NewCounterparty(other.ID, "Cafe"))
	require.NoError(t, err, "the same name is allowed for a different user")
}

func TestCounterpartyRepository_Create_UnknownUser(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)

	_, err := repo.Create(ctx, model.NewCounterparty(uuid.New(), "Cafe"))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "FK on user_id must reject unknown user, got: %v", err)
}

func TestCounterpartyRepository_DeleteByID_ClearsTransactionLink(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	counterparty := createTestCounterparty(t, ctx, user.ID)

	tx, err := transactions.Create(ctx, model.NewExpenseTransaction(account.ID, nil, &counterparty.ID, dec("5"), nil, testDate(0), true))
	require.NoError(t, err)

	require.NoError(t, repo.DeleteByID(ctx, counterparty.ID))

	exists, err := repo.ExistsByID(ctx, counterparty.ID)
	require.NoError(t, err)
	require.False(t, exists)

	fetched, err := transactions.GetByID(ctx, tx.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched, "deleting a counterparty must not delete its transactions")
	require.Nil(t, fetched.CounterpartyID, "ON DELETE SET NULL must clear the link")
	require.Equal(t, enum.TransactionExpense, fetched.Type)
}

func TestTransactionRepository_Create_UnknownCounterparty(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	ghost := uuid.New()

	_, err := repo.Create(ctx, model.NewExpenseTransaction(account.ID, nil, &ghost, dec("5"), nil, testDate(0), true))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "FK on counterparty_id must reject unknown counterparty, got: %v", err)
}

func TestCounterpartyRepository_ExistsByUserIDAndName(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	_, err := repo.Create(ctx, model.NewCounterparty(user.ID, "Cafe"))
	require.NoError(t, err)

	t.Run("same user and name", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndName(ctx, user.ID, "Cafe")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("different user", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndName(ctx, other.ID, "Cafe")
		require.NoError(t, err)
		require.False(t, exists, "another user's counterparty must not be visible")
	})

	t.Run("different name", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndName(ctx, user.ID, "Shop")
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func TestCounterpartyRepository_GetByUserIDAndName(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewCounterpartyRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	created, err := repo.Create(ctx, model.NewCounterparty(user.ID, "Cafe"))
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		found, err := repo.GetByUserIDAndName(ctx, user.ID, "Cafe")
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, created.ID, found.ID)
		require.Equal(t, "Cafe", found.Name)
	})

	t.Run("not found returns nil, nil", func(t *testing.T) {
		found, err := repo.GetByUserIDAndName(ctx, user.ID, "Missing")
		require.NoError(t, err)
		require.Nil(t, found)
	})

	t.Run("another user's counterparty is not reachable", func(t *testing.T) {
		found, err := repo.GetByUserIDAndName(ctx, other.ID, "Cafe")
		require.NoError(t, err)
		require.Nil(t, found)
	})
}
