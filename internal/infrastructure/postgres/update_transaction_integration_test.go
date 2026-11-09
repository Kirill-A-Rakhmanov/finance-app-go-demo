//go:build integration

package postgres_test

import (
	"context"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	pkgtx "finance-app-go-demo/pkg/transaction"

	"github.com/stretchr/testify/require"
)

func TestUpdateTransaction_KeepsBalancesNonNegativeAgainstTheDatabase(t *testing.T) {
	ctx := context.Background()
	accounts := postgres.NewAccountRepository(testClient)
	transactions := postgres.NewTransactionRepository(testClient)
	service := transaction.NewService(
		transactions,
		accounts,
		postgres.NewTransactionCategoryRepository(testClient),
		postgres.NewCounterpartyRepository(testClient),
		pkgtx.NewTransactionManager(testClient),
		nil,
	)

	user := createTestUser(t, ctx)
	source := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	target := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	foreign := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))

	_, err := transactions.Create(ctx, model.NewIncomeTransaction(source.ID, nil, nil, dec("100"), nil, testDate(0), true))
	require.NoError(t, err)
	spent, err := transactions.Create(ctx, model.NewExpenseTransaction(source.ID, nil, nil, dec("70"), nil, testDate(0), true))
	require.NoError(t, err)

	balanceOf := func(id model.Account) string {
		balance, err := transactions.GetBalanceByAccountID(ctx, id.ID)
		require.NoError(t, err)

		return balance.String()
	}
	update := func(accountID model.Account, type_ enum.TransactionType, amount string) error {
		_, err := service.UpdateTransaction(ctx, transaction.UpdateTransactionInput{
			TransactionID: spent.ID, UserID: user.ID, AccountID: accountID.ID,
			Type: type_, Amount: dec(amount), Date: testDate(0), IsConfirmed: true,
		})

		return err
	}

	require.Equal(t, "30", balanceOf(*source))

	require.NoError(t, update(*source, enum.TransactionExpense, "100"), "spending exactly the remaining 30 more is fine")
	require.Equal(t, "0", balanceOf(*source))

	require.ErrorIs(t, update(*source, enum.TransactionExpense, "101"), apperror.ErrNegativeBalanceNotAllowed)
	require.Equal(t, "0", balanceOf(*source), "a refused update changes nothing")

	require.NoError(t, update(*source, enum.TransactionExpense, "40"), "lowering an expense works even with nothing left")
	require.Equal(t, "60", balanceOf(*source))

	require.ErrorIs(t, update(*foreign, enum.TransactionExpense, "40"), transaction.ErrMoveToOtherCurrencyNotAllowed)

	require.ErrorIs(t, update(*target, enum.TransactionExpense, "40"), apperror.ErrNegativeBalanceNotAllowed, "the target account has no money for it")
	require.Equal(t, "60", balanceOf(*source))
	require.Equal(t, "0", balanceOf(*target))

	_, err = transactions.Create(ctx, model.NewIncomeTransaction(target.ID, nil, nil, dec("50"), nil, testDate(0), true))
	require.NoError(t, err)

	require.NoError(t, update(*target, enum.TransactionExpense, "40"), "now the target can afford it")
	require.Equal(t, "100", balanceOf(*source), "the expense left the source account")
	require.Equal(t, "10", balanceOf(*target))

	moved, err := transactions.GetByID(ctx, spent.ID)
	require.NoError(t, err)
	require.Equal(t, target.ID, moved.AccountID)
}
