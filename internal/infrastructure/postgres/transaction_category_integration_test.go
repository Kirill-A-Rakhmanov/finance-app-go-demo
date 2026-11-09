//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTransactionCategoryRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	created, err := repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Food", "fork", "#ff0000", enum.Expense))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, "Food", fetched.Name)
	require.Equal(t, "fork", fetched.Icon)
	require.Equal(t, "#ff0000", fetched.Color)
	require.Equal(t, enum.Expense, fetched.Type)
	require.Equal(t, user.ID, fetched.UserID)
	require.Nil(t, fetched.ParentID)
	require.Nil(t, fetched.Slug)
	require.False(t, fetched.IsDefault)
}

func TestTransactionCategoryRepository_Create_WithParentAndSlug(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	parent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	slug := "groceries"
	child := model.NewTransactionCategory(user.ID, &parent.ID, "Groceries", "cart", "#00ff00", enum.Expense)
	child.Slug = &slug
	child.IsDefault = true

	created, err := repo.Create(ctx, child)
	require.NoError(t, err)
	require.NotNil(t, created.ParentID)
	require.Equal(t, parent.ID, *created.ParentID)
	require.NotNil(t, created.Slug)
	require.Equal(t, "groceries", *created.Slug)
	require.True(t, created.IsDefault)
}

func TestTransactionCategoryRepository_Create_UnknownParent(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	missing := uuid.New()
	_, err := repo.Create(ctx, model.NewTransactionCategory(user.ID, &missing, "Orphan", "i", "#000000", enum.Expense))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "FK on parent_id must reject unknown parent, got: %v", err)
}

func TestTransactionCategoryRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)

	fetched, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestTransactionCategoryRepository_GetByIDAndUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	category := createTestCategory(t, ctx, user.ID, enum.Income, nil)

	found, err := repo.GetByIDAndUserID(ctx, category.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, category.ID, found.ID)

	notFound, err := repo.GetByIDAndUserID(ctx, category.ID, other.ID)
	require.NoError(t, err)
	require.Nil(t, notFound, "another user's category must not be reachable")
}

func TestTransactionCategoryRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	first := createTestCategory(t, ctx, user.ID, enum.Income, nil)
	second := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	createTestCategory(t, ctx, other.ID, enum.Expense, nil)

	categories, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, categories, 2)

	ids := []uuid.UUID{categories[0].ID, categories[1].ID}
	require.ElementsMatch(t, []uuid.UUID{first.ID, second.ID}, ids)

	empty, err := repo.GetByUserID(ctx, uuid.New())
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestTransactionCategoryRepository_GetAllByParentID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	parent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	childA := createTestCategory(t, ctx, user.ID, enum.Expense, &parent.ID)
	childB := createTestCategory(t, ctx, user.ID, enum.Expense, &parent.ID)
	createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	children, err := repo.GetAllByParentID(ctx, parent.ID)
	require.NoError(t, err)
	require.Len(t, children, 2)
	require.ElementsMatch(t, []uuid.UUID{childA.ID, childB.ID}, []uuid.UUID{children[0].ID, children[1].ID})

	none, err := repo.GetAllByParentID(ctx, childA.ID)
	require.NoError(t, err)
	require.Empty(t, none)
}

func TestTransactionCategoryRepository_ExistsByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	category := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	exists, err := repo.ExistsByID(ctx, category.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByID(ctx, uuid.New())
	require.NoError(t, err)
	require.False(t, exists)
}

func TestTransactionCategoryRepository_BulkCreate(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	parentID := uuid.New()
	childID := uuid.New()
	slug := "parent"

	parent := model.NewTransactionCategory(user.ID, nil, "Parent", "i", "#111111", enum.Expense)
	parent.ID = parentID
	parent.Slug = &slug
	parent.IsDefault = true

	child := model.NewTransactionCategory(user.ID, &parentID, "Child", "i", "#222222", enum.Expense)
	child.ID = childID

	created, err := repo.BulkCreate(ctx, []model.TransactionCategory{*parent, *child})
	require.NoError(t, err)
	require.Len(t, created, 2)

	fetchedChild, err := repo.GetByID(ctx, childID)
	require.NoError(t, err)
	require.NotNil(t, fetchedChild)
	require.NotNil(t, fetchedChild.ParentID)
	require.Equal(t, parentID, *fetchedChild.ParentID)

	fetchedParent, err := repo.GetByID(ctx, parentID)
	require.NoError(t, err)
	require.True(t, fetchedParent.IsDefault)
	require.NotNil(t, fetchedParent.Slug)
}

func TestTransactionCategoryRepository_BulkCreate_Empty(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)

	created, err := repo.BulkCreate(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestTransactionCategoryRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	parent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	created := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	slug := "renamed"
	created.Name = "Renamed"
	created.Icon = "new-icon"
	created.Color = "#abcdef"
	created.ParentID = &parent.ID
	created.Slug = &slug

	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.Equal(t, "Renamed", updated.Name)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "Renamed", fetched.Name)
	require.Equal(t, "new-icon", fetched.Icon)
	require.Equal(t, "#abcdef", fetched.Color)
	require.NotNil(t, fetched.ParentID)
	require.Equal(t, parent.ID, *fetched.ParentID)
	require.NotNil(t, fetched.Slug)

	// nil parent / slug must clear the columns
	fetched.ParentID = nil
	fetched.Slug = nil
	_, err = repo.Update(ctx, fetched)
	require.NoError(t, err)

	cleared, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Nil(t, cleared.ParentID)
	require.Nil(t, cleared.Slug)
}

func TestTransactionCategoryRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	ghost := model.NewTransactionCategory(user.ID, nil, "Ghost", "i", "#000000", enum.Expense)
	ghost.ID = uuid.New()

	_, err := repo.Update(ctx, ghost)
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err))
}

func TestTransactionCategoryRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	category := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	require.NoError(t, repo.DeleteByID(ctx, category.ID))

	fetched, err := repo.GetByID(ctx, category.ID)
	require.NoError(t, err)
	require.Nil(t, fetched)

	// deleting a missing row is not an error
	require.NoError(t, repo.DeleteByID(ctx, category.ID))
}

func TestTransactionCategoryRepository_DeleteByID_CascadesToChildren(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)

	parent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	child := createTestCategory(t, ctx, user.ID, enum.Expense, &parent.ID)

	require.NoError(t, repo.DeleteByID(ctx, parent.ID))

	fetched, err := repo.GetByID(ctx, child.ID)
	require.NoError(t, err)
	require.Nil(t, fetched, "child categories cascade with their parent")
}

func TestTransactionCategoryRepository_DeleteByID_KeepsTransactions(t *testing.T) {
	ctx := context.Background()
	categories := postgres.NewTransactionCategoryRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, "USD", dec("0"))
	category := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	tx := createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), testDate(0), true, &category.ID)

	require.NoError(t, categories.DeleteByID(ctx, category.ID))

	fetched, err := transactions.GetByID(ctx, tx.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched, "transactions survive category deletion")
	require.Nil(t, fetched.TransactionCategoryID, "category reference is set to NULL")
}

func TestTransactionCategoryRepository_GetByIDsAndUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	first := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	second := createTestCategory(t, ctx, user.ID, enum.Income, nil)
	createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	foreign := createTestCategory(t, ctx, other.ID, enum.Expense, nil)

	t.Run("returns only the requested categories", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{first.ID, second.ID}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 2)
		require.ElementsMatch(t, []uuid.UUID{first.ID, second.ID}, []uuid.UUID{found[0].ID, found[1].ID})
	})

	t.Run("another user's category is not returned", func(t *testing.T) {
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

	t.Run("duplicate ids return the category once", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, []uuid.UUID{first.ID, first.ID}, user.ID)
		require.NoError(t, err)
		require.Len(t, found, 1)
	})

	t.Run("empty id list returns nothing", func(t *testing.T) {
		found, err := repo.GetByIDsAndUserID(ctx, nil, user.ID)
		require.NoError(t, err)
		require.Empty(t, found)
	})
}

func TestTransactionCategoryRepository_Create_DuplicateNamePerUserAndType(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	_, err := repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Gifts", "icon", "#ffffff", enum.Expense))
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Gifts", "icon", "#ffffff", enum.Expense))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "(user_id, type, name) must be unique, got: %v", err)

	_, err = repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Gifts", "icon", "#ffffff", enum.Income))
	require.NoError(t, err, "the same name is allowed for a different type")

	_, err = repo.Create(ctx, model.NewTransactionCategory(other.ID, nil, "Gifts", "icon", "#ffffff", enum.Expense))
	require.NoError(t, err, "the same name is allowed for a different user")
}

func TestTransactionCategoryRepository_ExistsByUserIDAndTypeAndName(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	_, err := repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Gifts", "icon", "#ffffff", enum.Expense))
	require.NoError(t, err)

	t.Run("same user, type and name", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndTypeAndName(ctx, user.ID, enum.Expense, "Gifts")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("different type", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndTypeAndName(ctx, user.ID, enum.Income, "Gifts")
		require.NoError(t, err)
		require.False(t, exists, "the same name under another type is a different category")
	})

	t.Run("different user", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndTypeAndName(ctx, other.ID, enum.Expense, "Gifts")
		require.NoError(t, err)
		require.False(t, exists, "another user's category must not be visible")
	})

	t.Run("different name", func(t *testing.T) {
		exists, err := repo.ExistsByUserIDAndTypeAndName(ctx, user.ID, enum.Expense, "Gift")
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func TestTransactionCategoryRepository_GetByUserIDAndTypeAndName(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionCategoryRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	expense, err := repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Gifts", "icon", "#ffffff", enum.Expense))
	require.NoError(t, err)
	income, err := repo.Create(ctx, model.NewTransactionCategory(user.ID, nil, "Gifts", "icon", "#ffffff", enum.Income))
	require.NoError(t, err)

	t.Run("picks the category of the requested type", func(t *testing.T) {
		found, err := repo.GetByUserIDAndTypeAndName(ctx, user.ID, enum.Expense, "Gifts")
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, expense.ID, found.ID)

		found, err = repo.GetByUserIDAndTypeAndName(ctx, user.ID, enum.Income, "Gifts")
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, income.ID, found.ID)
	})

	t.Run("not found returns nil, nil", func(t *testing.T) {
		found, err := repo.GetByUserIDAndTypeAndName(ctx, user.ID, enum.Expense, "Missing")
		require.NoError(t, err)
		require.Nil(t, found)
	})

	t.Run("another user's category is not reachable", func(t *testing.T) {
		found, err := repo.GetByUserIDAndTypeAndName(ctx, other.ID, enum.Expense, "Gifts")
		require.NoError(t, err)
		require.Nil(t, found)
	})
}
