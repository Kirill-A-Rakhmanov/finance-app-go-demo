//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCounterpartyReportCriteria(userID uuid.UUID) query.GetCounterpartyReportCriteria {
	return query.GetCounterpartyReportCriteria{
		UserID:          userID,
		TransactionType: enum.TransactionExpense,
		Currency:        valueObject.CurrencyUSD,
	}
}

func counterpartyReportAmounts(result *query.GetCounterpartyReportResult) map[uuid.UUID]string {
	amounts := make(map[uuid.UUID]string, len(result.Items))
	for _, item := range result.Items {
		if item.Counterparty == nil {
			continue
		}

		amounts[item.Counterparty.ID] = item.Amount.String()
	}

	return amounts
}

func withoutCounterpartyAmount(t *testing.T, result *query.GetCounterpartyReportResult) decimal.Decimal {
	t.Helper()

	amount := decimal.Zero
	found := 0

	for _, item := range result.Items {
		if item.Counterparty == nil {
			amount = item.Amount
			found++
		}
	}

	require.LessOrEqual(t, found, 1, "all transactions without a counterparty must be one item")

	return amount
}

func createTestCounterpartyTransaction(
	t *testing.T,
	ctx context.Context,
	accountID uuid.UUID,
	type_ enum.TransactionType,
	amount decimal.Decimal,
	date time.Time,
	isConfirmed bool,
	counterpartyID *uuid.UUID,
) *model.Transaction {
	t.Helper()

	created, err := postgres.NewTransactionRepository(testClient).Create(
		ctx,
		model.NewTransaction(accountID, nil, counterpartyID, nil, type_, amount, nil, date, isConfirmed),
	)
	require.NoError(t, err)

	return created
}

func TestReportQueryService_GetCounterpartyReport_GroupsAndSortsByAmount(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	cafe := createTestCounterparty(t, ctx, user.ID)
	shop := createTestCounterparty(t, ctx, user.ID)
	gym := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(3), true, &cafe.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("15.5"), testDate(2), true, &cafe.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("100"), testDate(2), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("20"), testDate(1), true, &gym.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("7"), testDate(1), true, nil)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("3"), testDate(0), true, nil)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 4)
	assert.Equal(t, shop.ID, result.Items[0].Counterparty.ID)
	assert.True(t, result.Items[0].Amount.Equal(dec("100")), "got %s", result.Items[0].Amount)
	assert.Equal(t, cafe.ID, result.Items[1].Counterparty.ID)
	assert.True(t, result.Items[1].Amount.Equal(dec("25.5")), "got %s", result.Items[1].Amount)
	assert.Equal(t, gym.ID, result.Items[2].Counterparty.ID)
	assert.True(t, result.Items[2].Amount.Equal(dec("20")), "got %s", result.Items[2].Amount)
	assert.Nil(t, result.Items[3].Counterparty, "transactions without a counterparty come as the last, smallest item")
	assert.True(t, result.Items[3].Amount.Equal(dec("10")), "got %s", result.Items[3].Amount)
}

func TestReportQueryService_GetCounterpartyReport_TieBreaksByCounterpartyID(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	a := createTestCounterparty(t, ctx, user.ID)
	b := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), testDate(0), true, &a.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), testDate(0), true, &b.ID)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 2)

	first, second := a.ID, b.ID
	if second.String() < first.String() {
		first, second = second, first
	}
	assert.Equal(t, first, result.Items[0].Counterparty.ID)
	assert.Equal(t, second, result.Items[1].Counterparty.ID)
}

func TestReportQueryService_GetCounterpartyReport_EmptyReport(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	assert.NotNil(t, result.Items)
	assert.Empty(t, result.Items)
	assert.True(t, withoutCounterpartyAmount(t, result).IsZero())
}

func TestReportQueryService_GetCounterpartyReport_OnlyWithoutCounterparty(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("12"), testDate(0), true, nil)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Nil(t, result.Items[0].Counterparty)
	assert.True(t, result.Items[0].Amount.Equal(dec("12")))
}

func TestReportQueryService_GetCounterpartyReport_WithoutCounterpartyIsSortedByAmount(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	cafe := createTestCounterparty(t, ctx, user.ID)
	shop := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &cafe.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("50"), testDate(0), true, nil)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("100"), testDate(0), true, &shop.ID)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 3)
	assert.Equal(t, shop.ID, result.Items[0].Counterparty.ID)
	assert.Nil(t, result.Items[1].Counterparty, "the item without a counterparty sits between the others by amount")
	assert.True(t, result.Items[1].Amount.Equal(dec("50")), "got %s", result.Items[1].Amount)
	assert.Equal(t, cafe.ID, result.Items[2].Counterparty.ID)
}

func TestReportQueryService_GetCounterpartyReport_TypeFilter(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)
	employer := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("500"), testDate(0), true, &employer.ID)

	criteria := newCounterpartyReportCriteria(user.ID)
	criteria.TransactionType = enum.TransactionIncome

	result, err := qs.GetCounterpartyReport(ctx, criteria)

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{employer.ID: "500"}, counterpartyReportAmounts(result))
}

func TestReportQueryService_GetCounterpartyReport_OnlyConfirmedTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("99"), testDate(0), false, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("77"), testDate(0), false, nil)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "10"}, counterpartyReportAmounts(result))
	assert.True(t, withoutCounterpartyAmount(t, result).IsZero())
}

func TestReportQueryService_GetCounterpartyReport_ExcludesTransferTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, from.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &shop.ID)

	transfer, err := postgres.NewTransferRepository(testClient).Create(ctx, model.NewTransfer(from.ID, to.ID, dec("300"), dec("300"), testDate(0)))
	require.NoError(t, err)
	_, err = postgres.NewTransactionRepository(testClient).Create(ctx, model.NewTransaction(
		from.ID, nil, nil, &transfer.ID, enum.TransactionExpense, dec("300"), nil, testDate(0), true,
	))
	require.NoError(t, err)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "10"}, counterpartyReportAmounts(result))
	assert.True(t, withoutCounterpartyAmount(t, result).IsZero(), "transfer leg must not land in the 'no counterparty' bucket")
}

func TestReportQueryService_GetCounterpartyReport_CurrencySeparation(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	usd := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	eur := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, usd.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, eur.ID, enum.TransactionExpense, dec("400"), testDate(0), true, &shop.ID)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "10"}, counterpartyReportAmounts(result))

	criteria := newCounterpartyReportCriteria(user.ID)
	criteria.Currency = valueObject.CurrencyEUR
	result, err = qs.GetCounterpartyReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "400"}, counterpartyReportAmounts(result))
}

func TestReportQueryService_GetCounterpartyReport_IsolatesUsers(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	otherAccount := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)
	otherShop := createTestCounterparty(t, ctx, other.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, otherAccount.ID, enum.TransactionExpense, dec("999"), testDate(0), true, &otherShop.ID)
	createTestCounterpartyTransaction(t, ctx, otherAccount.ID, enum.TransactionExpense, dec("888"), testDate(0), true, nil)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "10"}, counterpartyReportAmounts(result))
	assert.True(t, withoutCounterpartyAmount(t, result).IsZero())
}

func TestReportQueryService_GetCounterpartyReport_PeriodBoundsAreInclusive(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(10), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("2"), testDate(5), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("4"), testDate(3), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("8"), testDate(1), true, &shop.ID)

	start, end := testDate(5), testDate(3)

	criteria := newCounterpartyReportCriteria(user.ID)
	criteria.PeriodStart = &start
	criteria.PeriodEnd = &end
	result, err := qs.GetCounterpartyReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "6"}, counterpartyReportAmounts(result))

	criteria = newCounterpartyReportCriteria(user.ID)
	criteria.PeriodStart = &start
	result, err = qs.GetCounterpartyReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "14"}, counterpartyReportAmounts(result))

	criteria = newCounterpartyReportCriteria(user.ID)
	criteria.PeriodEnd = &end
	result, err = qs.GetCounterpartyReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{shop.ID: "7"}, counterpartyReportAmounts(result))
}

func TestReportQueryService_GetCounterpartyReport_AccountsFilter(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	regular := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	another := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	currency, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)
	excluded, err := postgres.NewAccountRepository(testClient).Create(ctx,
		model.NewAccount(user.ID, "Excluded", "icon", currency, false, false, true, dec("0")))
	require.NoError(t, err)
	shop := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, regular.ID, enum.TransactionExpense, dec("1"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, another.ID, enum.TransactionExpense, dec("2"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, excluded.ID, enum.TransactionExpense, dec("4"), testDate(0), true, &shop.ID)

	t.Run("no accounts means all accounts except excluded ones", func(t *testing.T) {
		result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{shop.ID: "3"}, counterpartyReportAmounts(result))
	})

	t.Run("explicit accounts restrict the report", func(t *testing.T) {
		criteria := newCounterpartyReportCriteria(user.ID)
		criteria.AccountIDs = []uuid.UUID{another.ID}

		result, err := qs.GetCounterpartyReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{shop.ID: "2"}, counterpartyReportAmounts(result))
	})

	t.Run("an excluded account is included when requested explicitly", func(t *testing.T) {
		criteria := newCounterpartyReportCriteria(user.ID)
		criteria.AccountIDs = []uuid.UUID{regular.ID, excluded.ID}

		result, err := qs.GetCounterpartyReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{shop.ID: "5"}, counterpartyReportAmounts(result))
	})
}

func TestReportQueryService_GetCounterpartyReport_CounterpartyFilter(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	cafe := createTestCounterparty(t, ctx, user.ID)
	shop := createTestCounterparty(t, ctx, user.ID)
	gym := createTestCounterparty(t, ctx, user.ID)

	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &cafe.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("20"), testDate(0), true, &shop.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("30"), testDate(0), true, &gym.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("40"), testDate(0), true, nil)

	t.Run("only the requested counterparties are returned", func(t *testing.T) {
		criteria := newCounterpartyReportCriteria(user.ID)
		criteria.CounterpartyIDs = []uuid.UUID{cafe.ID, gym.ID}

		result, err := qs.GetCounterpartyReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{cafe.ID: "10", gym.ID: "30"}, counterpartyReportAmounts(result))
	})

	t.Run("the filter drops transactions without a counterparty", func(t *testing.T) {
		criteria := newCounterpartyReportCriteria(user.ID)
		criteria.CounterpartyIDs = []uuid.UUID{cafe.ID}

		result, err := qs.GetCounterpartyReport(ctx, criteria)

		require.NoError(t, err)
		assert.True(t, withoutCounterpartyAmount(t, result).IsZero())
	})

	t.Run("filter of another type yields an empty report", func(t *testing.T) {
		criteria := newCounterpartyReportCriteria(user.ID)
		criteria.TransactionType = enum.TransactionIncome
		criteria.CounterpartyIDs = []uuid.UUID{cafe.ID}

		result, err := qs.GetCounterpartyReport(ctx, criteria)

		require.NoError(t, err)
		assert.Empty(t, result.Items)
	})
}

func TestReportQueryService_GetCounterpartyReport_MapsCounterpartyFields(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	shop := createTestCounterparty(t, ctx, user.ID)
	createTestCounterpartyTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(0), true, &shop.ID)

	result, err := qs.GetCounterpartyReport(ctx, newCounterpartyReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	got := result.Items[0].Counterparty
	assert.Equal(t, shop.ID, got.ID)
	assert.Equal(t, user.ID, got.UserID)
	assert.Equal(t, shop.Name, got.Name)
	assert.False(t, got.CreatedAt.IsZero())
}
