//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestEmail(t *testing.T) valueObject.Email {
	t.Helper()

	email, err := valueObject.NewEmail(fmt.Sprintf("user-%s@example.com", uuid.NewString()))
	require.NoError(t, err)

	return email
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)

	email := newTestEmail(t)
	middle := "Q"

	created, err := repo.Create(ctx, model.NewUser("Jane", "Doe", email, "hash", &middle))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)
	require.False(t, created.CreatedAt.IsZero())

	byID, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, byID)
	require.Equal(t, "Jane", byID.FirstName)
	require.Equal(t, "Doe", byID.LastName)
	require.Equal(t, email.ToString(), byID.Email.ToString())
	require.Equal(t, "hash", byID.Password)
	require.NotNil(t, byID.MiddleName)
	require.Equal(t, "Q", *byID.MiddleName)

	byEmail, err := repo.GetByEmail(ctx, email.ToString())
	require.NoError(t, err)
	require.NotNil(t, byEmail)
	require.Equal(t, created.ID, byEmail.ID)
}

func TestUserRepository_Create_NilMiddleName(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)

	created, err := repo.Create(ctx, model.NewUser("Jane", "Doe", newTestEmail(t), "hash", nil))
	require.NoError(t, err)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Nil(t, fetched.MiddleName)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)

	email := newTestEmail(t)

	_, err := repo.Create(ctx, model.NewUser("Jane", "Doe", email, "hash", nil))
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.NewUser("John", "Roe", email, "hash", nil))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "duplicate email must violate the unique constraint, got: %v", err)
}

func TestUserRepository_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)

	byID, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, byID)

	byEmail, err := repo.GetByEmail(ctx, "missing-"+uuid.NewString()+"@example.com")
	require.NoError(t, err)
	require.Nil(t, byEmail)
}

func TestUserRepository_Exists(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)
	user := createTestUser(t, ctx)

	exists, err := repo.ExistsByID(ctx, user.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByID(ctx, uuid.New())
	require.NoError(t, err)
	require.False(t, exists)

	exists, err = repo.ExistsByEmail(ctx, user.Email.ToString())
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByEmail(ctx, "missing-"+uuid.NewString()+"@example.com")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)

	middle := "Q"
	created, err := repo.Create(ctx, model.NewUser("Jane", "Doe", newTestEmail(t), "hash", &middle))
	require.NoError(t, err)

	newEmail := newTestEmail(t)
	created.FirstName = "Janet"
	created.LastName = "Smith"
	created.Email = newEmail
	created.Password = "new-hash"
	created.MiddleName = nil

	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, "Janet", updated.FirstName)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "Janet", fetched.FirstName)
	require.Equal(t, "Smith", fetched.LastName)
	require.Equal(t, newEmail.ToString(), fetched.Email.ToString())
	require.Equal(t, "new-hash", fetched.Password)
	require.Nil(t, fetched.MiddleName, "nil MiddleName must clear the column")
}

func TestUserRepository_Update_EmailTakenByOther(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)
	first := createTestUser(t, ctx)
	second := createTestUser(t, ctx)

	second.Email = first.Email

	_, err := repo.Update(ctx, second)
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err))
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewUserRepository(testClient)

	ghost := model.NewUser("No", "Body", newTestEmail(t), "hash", nil)
	ghost.ID = uuid.New()

	_, err := repo.Update(ctx, ghost)
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err))
}

func TestUserRepository_DeleteByID_CascadesOwnedData(t *testing.T) {
	ctx := context.Background()
	users := postgres.NewUserRepository(testClient)
	accounts := postgres.NewAccountRepository(testClient)
	categories := postgres.NewTransactionCategoryRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)

	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	category := createTestCategory(t, ctx, user.ID, "expense", nil)
	tx := createTestTransaction(t, ctx, account.ID, "expense", dec("10"), testDate(0), true, &category.ID)
	otherAccount := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, dec("0"))

	require.NoError(t, users.DeleteByID(ctx, user.ID))

	got, err := users.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Nil(t, got)

	gotAccount, err := accounts.GetByID(ctx, account.ID)
	require.NoError(t, err)
	require.Nil(t, gotAccount, "accounts must cascade on user delete")

	gotCategory, err := categories.GetByID(ctx, category.ID)
	require.NoError(t, err)
	require.Nil(t, gotCategory, "categories must cascade on user delete")

	gotTx, err := transactions.GetByID(ctx, tx.ID)
	require.NoError(t, err)
	require.Nil(t, gotTx, "transactions must cascade via account")

	gotOther, err := accounts.GetByID(ctx, otherAccount.ID)
	require.NoError(t, err)
	require.NotNil(t, gotOther, "another user's data must be untouched")
}
