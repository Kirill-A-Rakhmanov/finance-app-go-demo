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

func TestTransactionRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	category := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	counterparty := createTestCounterparty(t, ctx, user.ID)

	name := "Lunch"
	date := testDate(1)

	created, err := repo.Create(ctx, model.NewExpenseTransaction(account.ID, &category.ID, &counterparty.ID, dec("12.3456"), &name, date, true))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, account.ID, fetched.AccountID)
	require.NotNil(t, fetched.TransactionCategoryID)
	require.Equal(t, category.ID, *fetched.TransactionCategoryID)
	require.Nil(t, fetched.TransferID)
	require.Equal(t, enum.TransactionExpense, fetched.Type)
	require.True(t, fetched.Amount.Equal(dec("12.3456")), "amount keeps 4 decimal places, got %s", fetched.Amount)
	require.Equal(t, "Lunch", *fetched.Name)
	require.NotNil(t, fetched.CounterpartyID)
	require.Equal(t, counterparty.ID, *fetched.CounterpartyID)
	require.True(t, fetched.Date.Equal(date))
	require.True(t, fetched.IsConfirmed)
}

func TestTransactionRepository_Create_OptionalFieldsNil(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	created := createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1"), testDate(0), false, nil)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Nil(t, fetched.TransactionCategoryID)
	require.Nil(t, fetched.Name)
	require.Nil(t, fetched.CounterpartyID)
	require.False(t, fetched.IsConfirmed)
}

func TestTransactionRepository_Create_UnknownAccount(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)

	_, err := repo.Create(ctx, model.NewIncomeTransaction(uuid.New(), nil, nil, dec("1"), nil, testDate(0), true))
	require.Error(t, err)
	require.True(t, ent.IsConstraintError(err), "FK on account_id must reject unknown account, got: %v", err)
}

func TestTransactionRepository_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)

	fetched, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestTransactionRepository_ExistsByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	tx := createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)

	exists, err := repo.ExistsByID(ctx, tx.ID)
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = repo.ExistsByID(ctx, uuid.New())
	require.NoError(t, err)
	require.False(t, exists)
}

func TestTransactionRepository_GetByIDAndUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	tx := createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)

	found, err := repo.GetByIDAndUserID(ctx, tx.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	notFound, err := repo.GetByIDAndUserID(ctx, tx.ID, other.ID)
	require.NoError(t, err)
	require.Nil(t, notFound, "another user's transaction must not be reachable")
}

func TestTransactionRepository_GetByIDAndAccountID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	otherAccount := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	tx := createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)

	found, err := repo.GetByIDAndAccountID(ctx, tx.ID, account.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	notFound, err := repo.GetByIDAndAccountID(ctx, tx.ID, otherAccount.ID)
	require.NoError(t, err)
	require.Nil(t, notFound)
}

func TestTransactionRepository_GetByAccountID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	otherAccount := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	a := createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)
	b := createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("2"), testDate(1), true, nil)
	createTestTransaction(t, ctx, otherAccount.ID, enum.TransactionIncome, dec("3"), testDate(0), true, nil)

	txs, err := repo.GetByAccountID(ctx, account.ID)
	require.NoError(t, err)
	require.Len(t, txs, 2)
	require.ElementsMatch(t, []uuid.UUID{a.ID, b.ID}, []uuid.UUID{txs[0].ID, txs[1].ID})
}

func TestTransactionRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	accountA := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	accountB := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))
	otherAccount := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, dec("0"))

	a := createTestTransaction(t, ctx, accountA.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)
	b := createTestTransaction(t, ctx, accountB.ID, enum.TransactionIncome, dec("2"), testDate(0), true, nil)
	createTestTransaction(t, ctx, otherAccount.ID, enum.TransactionIncome, dec("3"), testDate(0), true, nil)

	txs, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, txs, 2)
	require.ElementsMatch(t, []uuid.UUID{a.ID, b.ID}, []uuid.UUID{txs[0].ID, txs[1].ID})
}

func TestTransactionRepository_GetByTransactionCategoryID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	category := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	otherCategory := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	in := createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(0), true, &category.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(0), true, &otherCategory.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(0), true, nil)

	txs, err := repo.GetByTransactionCategoryID(ctx, category.ID)
	require.NoError(t, err)
	require.Len(t, txs, 1)
	require.Equal(t, in.ID, txs[0].ID)
}

func TestTransactionRepository_BulkCreate(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	created, err := repo.BulkCreate(ctx, []model.Transaction{
		*model.NewIncomeTransaction(account.ID, nil, nil, dec("10"), nil, testDate(0), true),
		*model.NewExpenseTransaction(account.ID, nil, nil, dec("4"), nil, testDate(0), true),
	})
	require.NoError(t, err)
	require.Len(t, created, 2)
	require.NotEqual(t, created[0].ID, created[1].ID)

	txs, err := repo.GetByAccountID(ctx, account.ID)
	require.NoError(t, err)
	require.Len(t, txs, 2)
}

func TestTransactionRepository_BulkCreate_Empty(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)

	created, err := repo.BulkCreate(ctx, nil)
	require.NoError(t, err)
	require.Empty(t, created)
}

func TestTransactionRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	otherAccount := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	category := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	counterparty := createTestCounterparty(t, ctx, user.ID)
	name := "Old"
	created, err := repo.Create(ctx, model.NewExpenseTransaction(account.ID, &category.ID, &counterparty.ID, dec("5"), &name, testDate(3), false))
	require.NoError(t, err)

	newName := "New"
	created.AccountID = otherAccount.ID
	created.Type = enum.TransactionIncome
	created.Amount = dec("99.5")
	created.Name = &newName
	created.CounterpartyID = nil
	created.TransactionCategoryID = nil
	created.Date = testDate(0)
	created.IsConfirmed = true

	_, err = repo.Update(ctx, created)
	require.NoError(t, err)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, otherAccount.ID, fetched.AccountID)
	require.Equal(t, enum.TransactionIncome, fetched.Type)
	require.True(t, fetched.Amount.Equal(dec("99.5")))
	require.Equal(t, "New", *fetched.Name)
	require.Nil(t, fetched.CounterpartyID, "nil counterparty must clear the column")
	require.Nil(t, fetched.TransactionCategoryID, "nil category must clear the column")
	require.True(t, fetched.Date.Equal(testDate(0)))
	require.True(t, fetched.IsConfirmed)
}

func TestTransactionRepository_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	ghost := model.NewIncomeTransaction(account.ID, nil, nil, dec("1"), nil, testDate(0), true)
	ghost.ID = uuid.New()

	_, err := repo.Update(ctx, ghost)
	require.Error(t, err)
	require.True(t, ent.IsNotFound(err))
}

func TestTransactionRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	tx := createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)

	require.NoError(t, repo.DeleteByID(ctx, tx.ID))

	fetched, err := repo.GetByID(ctx, tx.ID)
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestTransactionRepository_GetBalanceByAccountID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))

	createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("50.25"), testDate(2), true, nil)
	createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("10"), testDate(1), true, nil)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("20.75"), testDate(0), true, nil)
	// unconfirmed transactions must not affect the balance
	createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("1000"), testDate(0), false, nil)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("500"), testDate(0), false, nil)

	balance, err := repo.GetBalanceByAccountID(ctx, account.ID)
	require.NoError(t, err)
	// 100 (adjustment) + 50.25 + 10 - 20.75
	require.True(t, balance.Equal(dec("139.5")), "got %s", balance)
}

func TestTransactionRepository_GetBalanceByAccountID_NoTransactions(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("42"))

	balance, err := repo.GetBalanceByAccountID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, balance.Equal(dec("42")), "got %s", balance)
}

func TestTransactionRepository_GetBalanceByAccountID_OnlyOwnTransactions(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	otherAccount := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("10"), testDate(0), true, nil)
	createTestTransaction(t, ctx, otherAccount.ID, enum.TransactionIncome, dec("999"), testDate(0), true, nil)

	balance, err := repo.GetBalanceByAccountID(ctx, account.ID)
	require.NoError(t, err)
	require.True(t, balance.Equal(dec("10")), "got %s", balance)
}

func TestTransactionRepository_GetBalanceByUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)

	usdA := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("10"))
	usdB := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	eur := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))
	otherUsd := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, dec("0"))

	createTestTransaction(t, ctx, usdA.ID, enum.TransactionIncome, dec("100"), testDate(0), true, nil)
	createTestTransaction(t, ctx, usdA.ID, enum.TransactionExpense, dec("30"), testDate(0), true, nil)
	createTestTransaction(t, ctx, usdB.ID, enum.TransactionIncome, dec("5"), testDate(0), true, nil)
	createTestTransaction(t, ctx, usdB.ID, enum.TransactionIncome, dec("777"), testDate(0), false, nil)
	createTestTransaction(t, ctx, eur.ID, enum.TransactionIncome, dec("8"), testDate(0), true, nil)
	createTestTransaction(t, ctx, otherUsd.ID, enum.TransactionIncome, dec("1000"), testDate(0), true, nil)

	balances, err := repo.GetBalanceByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, balances, 2)

	// sorted by currency code: EUR, USD
	require.Equal(t, valueObject.CurrencyEUR, balances[0].Currency.Code)
	require.True(t, balances[0].Amount.Equal(dec("8")), "got %s", balances[0].Amount)
	require.Equal(t, valueObject.CurrencyUSD, balances[1].Currency.Code)
	// 10 + 100 - 30 + 5
	require.True(t, balances[1].Amount.Equal(dec("85")), "got %s", balances[1].Amount)
}

func TestTransactionRepository_GetBalanceByUserID_ExcludesFlaggedAccounts(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	accounts := postgres.NewAccountRepository(testClient)
	user := createTestUser(t, ctx)

	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	counted := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("10"))
	excluded, err := accounts.Create(ctx, model.NewAccount(user.ID, "Excluded", "icon", usd, false, false, true, dec("500")))
	require.NoError(t, err)
	createTestTransaction(t, ctx, excluded.ID, enum.TransactionIncome, dec("100"), testDate(0), true, nil)
	createTestTransaction(t, ctx, counted.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)

	balances, err := repo.GetBalanceByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, balances, 1)
	require.True(t, balances[0].Amount.Equal(dec("11")), "got %s", balances[0].Amount)
}

func TestTransactionRepository_GetBalanceByUserID_NoAccounts(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)

	balances, err := repo.GetBalanceByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, balances)
	require.Empty(t, balances)
}

func TestTransactionRepository_GetBalanceByUserID_AccountsWithoutTransactions(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewTransactionRepository(testClient)
	user := createTestUser(t, ctx)
	createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("7"))

	balances, err := repo.GetBalanceByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, balances, 1)
	require.True(t, balances[0].Amount.Equal(dec("7")), "got %s", balances[0].Amount)
}
