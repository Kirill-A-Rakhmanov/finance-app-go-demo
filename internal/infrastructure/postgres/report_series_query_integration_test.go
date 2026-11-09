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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

func at(year int, month time.Month, d, hour, minute int) time.Time {
	return time.Date(year, month, d, hour, minute, 0, 0, time.UTC)
}

func moscowLocation(t *testing.T) *time.Location {
	t.Helper()

	location, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)

	return location
}

func createSeriesTransaction(
	t *testing.T,
	ctx context.Context,
	accountID uuid.UUID,
	type_ enum.TransactionType,
	amount string,
	date time.Time,
	isConfirmed bool,
) {
	t.Helper()

	createTestCounterpartyTransaction(t, ctx, accountID, type_, dec(amount), date, isConfirmed, nil)
}

func createSeriesTransfer(
	t *testing.T,
	ctx context.Context,
	fromAccountID, toAccountID uuid.UUID,
	amount string,
	date time.Time,
) {
	t.Helper()

	transfer, err := postgres.NewTransferRepository(testClient).Create(
		ctx,
		model.NewTransfer(fromAccountID, toAccountID, dec(amount), dec(amount), date),
	)
	require.NoError(t, err)

	transactions := postgres.NewTransactionRepository(testClient)

	_, err = transactions.Create(ctx, model.NewTransaction(
		fromAccountID, nil, nil, &transfer.ID, enum.TransactionExpense, dec(amount), nil, date, true,
	))
	require.NoError(t, err)

	_, err = transactions.Create(ctx, model.NewTransaction(
		toAccountID, nil, nil, &transfer.ID, enum.TransactionIncome, dec(amount), nil, date, true,
	))
	require.NoError(t, err)
}

func createExcludedAccount(t *testing.T, ctx context.Context, userID uuid.UUID, adjustment string) *model.Account {
	t.Helper()

	currency, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	created, err := postgres.NewAccountRepository(testClient).Create(
		ctx,
		model.NewAccount(userID, "Excluded", "icon", currency, false, false, true, dec(adjustment)),
	)
	require.NoError(t, err)

	return created
}

func incomeExpenseCriteria(userID uuid.UUID, start, end time.Time) query.GetIncomeExpenseReportCriteria {
	return query.GetIncomeExpenseReportCriteria{
		UserID:      userID,
		Currency:    valueObject.CurrencyUSD,
		Location:    *time.UTC,
		PeriodStart: &start,
		PeriodEnd:   &end,
	}
}

func assertIncomeExpenseItem(t *testing.T, item query.IncomeExpenseReportItem, month, income, expense string) {
	t.Helper()

	assert.Equal(t, month, item.Month)
	assert.True(t, item.Income.Equal(dec(income)), "%s income: got %s, want %s", month, item.Income, income)
	assert.True(t, item.Expense.Equal(dec(expense)), "%s expense: got %s, want %s", month, item.Expense, expense)
}

func TestReportQueryService_GetIncomeExpenseReport_FillsEmptyMonthsWithZeros(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "1000", day(2026, 1, 10), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "100.25", day(2026, 1, 11), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "40", day(2026, 1, 31), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "500", day(2026, 4, 1), true)

	result, err := qs.GetIncomeExpenseReport(ctx, incomeExpenseCriteria(user.ID, day(2026, 1, 1), day(2026, 4, 30)))

	require.NoError(t, err)
	require.Len(t, result.Items, 4)
	assertIncomeExpenseItem(t, result.Items[0], "2026-01", "1000", "140.25")
	assertIncomeExpenseItem(t, result.Items[1], "2026-02", "0", "0")
	assertIncomeExpenseItem(t, result.Items[2], "2026-03", "0", "0")
	assertIncomeExpenseItem(t, result.Items[3], "2026-04", "500", "0")
}

func TestReportQueryService_GetIncomeExpenseReport_PeriodIsInclusiveByDate(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "1", at(2026, 3, 9, 23, 59), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "10", at(2026, 3, 10, 0, 0), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "100", at(2026, 3, 20, 23, 59), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "1000", at(2026, 3, 21, 0, 0), true)

	result, err := qs.GetIncomeExpenseReport(ctx, incomeExpenseCriteria(user.ID, day(2026, 3, 10), day(2026, 3, 20)))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assertIncomeExpenseItem(t, result.Items[0], "2026-03", "0", "110")
}

func TestReportQueryService_GetIncomeExpenseReport_MonthsFollowTheTimezone(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	// 01:30 on April 1st in Moscow (UTC+3), still March 31st in UTC
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "10", at(2026, 3, 31, 22, 30), true)

	utc := incomeExpenseCriteria(user.ID, day(2026, 3, 1), day(2026, 4, 30))
	result, err := qs.GetIncomeExpenseReport(ctx, utc)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assertIncomeExpenseItem(t, result.Items[0], "2026-03", "0", "10")
	assertIncomeExpenseItem(t, result.Items[1], "2026-04", "0", "0")

	moscow := incomeExpenseCriteria(user.ID, day(2026, 3, 1), day(2026, 4, 30))
	moscow.Location = *moscowLocation(t)
	result, err = qs.GetIncomeExpenseReport(ctx, moscow)
	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assertIncomeExpenseItem(t, result.Items[0], "2026-03", "0", "0")
	assertIncomeExpenseItem(t, result.Items[1], "2026-04", "0", "10")
}

func TestReportQueryService_GetIncomeExpenseReport_CountsOnlyConfirmedNonTransferTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	other := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	euro := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))
	stranger := createTestUser(t, ctx)
	strangerAccount := createTestAccount(t, ctx, stranger.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "10", day(2026, 3, 5), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "99", day(2026, 3, 5), false)
	createSeriesTransaction(t, ctx, euro.ID, enum.TransactionExpense, "77", day(2026, 3, 5), true)
	createSeriesTransaction(t, ctx, strangerAccount.ID, enum.TransactionExpense, "55", day(2026, 3, 5), true)
	createSeriesTransfer(t, ctx, account.ID, other.ID, "300", day(2026, 3, 5))

	result, err := qs.GetIncomeExpenseReport(ctx, incomeExpenseCriteria(user.ID, day(2026, 3, 1), day(2026, 3, 31)))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assertIncomeExpenseItem(t, result.Items[0], "2026-03", "0", "10")
}

func TestReportQueryService_GetIncomeExpenseReport_AccountFilter(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	counted := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	selected := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	excluded := createExcludedAccount(t, ctx, user.ID, "0")

	createSeriesTransaction(t, ctx, counted.ID, enum.TransactionExpense, "1", day(2026, 3, 5), true)
	createSeriesTransaction(t, ctx, selected.ID, enum.TransactionExpense, "10", day(2026, 3, 5), true)
	createSeriesTransaction(t, ctx, excluded.ID, enum.TransactionExpense, "100", day(2026, 3, 5), true)

	criteria := incomeExpenseCriteria(user.ID, day(2026, 3, 1), day(2026, 3, 31))

	result, err := qs.GetIncomeExpenseReport(ctx, criteria)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assertIncomeExpenseItem(t, result.Items[0], "2026-03", "0", "11")

	criteria.AccountIDs = []uuid.UUID{selected.ID, excluded.ID}
	result, err = qs.GetIncomeExpenseReport(ctx, criteria)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assertIncomeExpenseItem(t, result.Items[0], "2026-03", "0", "110")
}

func TestReportQueryService_GetIncomeExpenseReport_OpenPeriod(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	currentMonth := time.Now().UTC().Format("2006-01")
	firstMonth := time.Now().UTC().AddDate(0, -2, 0)
	firstMonth = time.Date(firstMonth.Year(), firstMonth.Month(), 15, 12, 0, 0, 0, time.UTC)

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "10", firstMonth, true)

	result, err := qs.GetIncomeExpenseReport(ctx, query.GetIncomeExpenseReportCriteria{
		UserID:   user.ID,
		Currency: valueObject.CurrencyUSD,
		Location: *time.UTC,
	})

	require.NoError(t, err)
	require.Len(t, result.Items, 3, "from the first transaction to the current month")
	assert.Equal(t, firstMonth.Format("2006-01"), result.Items[0].Month)
	assert.Equal(t, currentMonth, result.Items[2].Month)
	assertIncomeExpenseItem(t, result.Items[0], firstMonth.Format("2006-01"), "0", "10")
}

func TestReportQueryService_GetIncomeExpenseReport_NoTransactionsWithoutPeriodIsEmpty(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)

	result, err := qs.GetIncomeExpenseReport(ctx, query.GetIncomeExpenseReportCriteria{
		UserID:   user.ID,
		Currency: valueObject.CurrencyUSD,
		Location: *time.UTC,
	})

	require.NoError(t, err)
	assert.NotNil(t, result.Items)
	assert.Empty(t, result.Items)
}

func TestReportQueryService_GetIncomeExpenseReport_RequiresNamedLocation(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	tests := []struct {
		name     string
		location time.Location
	}{
		{"zero location", time.Location{}},
		{"local location has no name Postgres knows", *time.Local},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qs.GetIncomeExpenseReport(ctx, query.GetIncomeExpenseReportCriteria{
				UserID:   uuid.New(),
				Currency: valueObject.CurrencyUSD,
				Location: tt.location,
			})

			require.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func balanceCriteria(userID uuid.UUID, granularity query.Granularity, start, end time.Time) query.GetBalanceReportCriteria {
	return query.GetBalanceReportCriteria{
		UserID:      userID,
		Currency:    valueObject.CurrencyUSD,
		Location:    *time.UTC,
		Granularity: granularity,
		PeriodStart: start,
		PeriodEnd:   end,
	}
}

func balanceSeries(result *query.GetBalanceReportResult) map[string]string {
	series := make(map[string]string, len(result.Items))
	for _, item := range result.Items {
		series[item.DateKey] = item.Balance.String()
	}

	return series
}

func assertBalanceKeys(t *testing.T, result *query.GetBalanceReportResult, keys ...string) {
	t.Helper()

	got := make([]string, len(result.Items))
	for i, item := range result.Items {
		got[i] = item.DateKey
	}

	assert.Equal(t, keys, got)
}

func assertBalance(t *testing.T, item query.BalanceReportItem, want string) {
	t.Helper()

	assert.True(t, item.Balance.Equal(dec(want)), "%s: got %s, want %s", item.DateKey, item.Balance, want)
}

func TestReportQueryService_GetBalanceReport_DailyStartsFromOpeningBalance(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("1000"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "500", day(2026, 2, 1), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "200", day(2026, 2, 20), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "50", day(2026, 3, 3), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "10", day(2026, 3, 5), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "999", day(2026, 3, 6), true)

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityDay, day(2026, 3, 1), day(2026, 3, 5)))

	require.NoError(t, err)
	assertBalanceKeys(t, result, "2026-03-01", "2026-03-02", "2026-03-03", "2026-03-04", "2026-03-05")
	assertBalance(t, result.Items[0], "1300")
	assertBalance(t, result.Items[1], "1300")
	assertBalance(t, result.Items[2], "1250")
	assertBalance(t, result.Items[3], "1250")
	assertBalance(t, result.Items[4], "1260")
}

func TestReportQueryService_GetBalanceReport_WeeksStartOnMondayAndAreCutToThePeriod(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "100", day(2026, 6, 1), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "10", day(2026, 6, 3), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "1", day(2026, 6, 7), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "5", day(2026, 6, 8), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "1000", day(2026, 6, 30), true)

	// 2026-06-03 is a Wednesday, 2026-06-30 a Tuesday
	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityWeek, day(2026, 6, 3), day(2026, 6, 30)))

	require.NoError(t, err)
	assertBalanceKeys(t, result, "2026-06-07", "2026-06-14", "2026-06-21", "2026-06-28", "2026-06-30")
	assertBalance(t, result.Items[0], "111")
	assertBalance(t, result.Items[1], "106")
	assertBalance(t, result.Items[2], "106")
	assertBalance(t, result.Items[3], "106")
	assertBalance(t, result.Items[4], "1106")
}

func TestReportQueryService_GetBalanceReport_MonthsStartOnTheFirstAndAreCutToThePeriod(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "10", day(2026, 1, 10), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "20", day(2026, 1, 20), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "5", day(2026, 2, 28), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "1", day(2026, 3, 10), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "1000", day(2026, 3, 11), true)

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityMonth, day(2026, 1, 15), day(2026, 3, 10)))

	require.NoError(t, err)
	assertBalanceKeys(t, result, "2026-01-31", "2026-02-28", "2026-03-10")
	assertBalance(t, result.Items[0], "130")
	assertBalance(t, result.Items[1], "125")
	assertBalance(t, result.Items[2], "126")
}

func TestReportQueryService_GetBalanceReport_SingleBucketInsideOneMonth(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "30", day(2026, 5, 12), true)

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityMonth, day(2026, 5, 10), day(2026, 5, 20)))

	require.NoError(t, err)
	assertBalanceKeys(t, result, "2026-05-20")
	assertBalance(t, result.Items[0], "70")
}

func TestReportQueryService_GetBalanceReport_TransfersMoveMoneyOutOfCountedAccounts(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	main := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("1000"))
	savings := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	hidden := createExcludedAccount(t, ctx, user.ID, "5000")

	createSeriesTransfer(t, ctx, main.ID, savings.ID, "200", day(2026, 3, 2))
	createSeriesTransfer(t, ctx, main.ID, hidden.ID, "100", day(2026, 3, 3))

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityDay, day(2026, 3, 1), day(2026, 3, 3)))

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"2026-03-01": "1000",
		"2026-03-02": "1000", // main -200, savings +200
		"2026-03-03": "900",  // main -100, the receiving account is not counted
	}, balanceSeries(result))
}

func TestReportQueryService_GetBalanceReport_AccountFilterAndIgnoredTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	main := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))
	selected := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("10"))
	euro := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("7"))
	stranger := createTestUser(t, ctx)
	strangerAccount := createTestAccount(t, ctx, stranger.ID, valueObject.CurrencyUSD, dec("9999"))

	createSeriesTransaction(t, ctx, main.ID, enum.TransactionIncome, "1", day(2026, 3, 2), true)
	createSeriesTransaction(t, ctx, main.ID, enum.TransactionIncome, "50", day(2026, 3, 2), false)
	createSeriesTransaction(t, ctx, selected.ID, enum.TransactionIncome, "2", day(2026, 3, 2), true)
	createSeriesTransaction(t, ctx, euro.ID, enum.TransactionIncome, "500", day(2026, 3, 2), true)
	createSeriesTransaction(t, ctx, strangerAccount.ID, enum.TransactionIncome, "700", day(2026, 3, 2), true)

	criteria := balanceCriteria(user.ID, query.GranularityDay, day(2026, 3, 1), day(2026, 3, 2))

	result, err := qs.GetBalanceReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2026-03-01": "110", "2026-03-02": "113"}, balanceSeries(result))

	criteria.AccountIDs = []uuid.UUID{selected.ID}
	result, err = qs.GetBalanceReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2026-03-01": "10", "2026-03-02": "12"}, balanceSeries(result))
}

func TestReportQueryService_GetBalanceReport_DaysFollowTheTimezone(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	// 01:30 on March 2nd in Moscow (UTC+3), still March 1st in UTC
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "10", at(2026, 3, 1, 22, 30), true)

	criteria := balanceCriteria(user.ID, query.GranularityDay, day(2026, 3, 1), day(2026, 3, 2))

	result, err := qs.GetBalanceReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2026-03-01": "10", "2026-03-02": "10"}, balanceSeries(result))

	criteria.Location = *moscowLocation(t)
	result, err = qs.GetBalanceReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2026-03-01": "0", "2026-03-02": "10"}, balanceSeries(result))
}

func TestReportQueryService_GetBalanceReport_UnknownGranularity(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, "year", day(2026, 3, 1), day(2026, 3, 2)))

	require.Error(t, err)
	assert.Nil(t, result)
}

// pkg/postgres hands out a Debug() client, which wraps the driver; raw SQL has to survive that.
func TestReportQueryService_SeriesReportsWorkWithTheDebugClient(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient.Debug())

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "30", day(2026, 3, 2), true)

	balance, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityDay, day(2026, 3, 1), day(2026, 3, 2)))
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2026-03-01": "100", "2026-03-02": "70"}, balanceSeries(balance))

	incomeExpense, err := qs.GetIncomeExpenseReport(ctx, incomeExpenseCriteria(user.ID, day(2026, 3, 1), day(2026, 3, 31)))
	require.NoError(t, err)
	require.Len(t, incomeExpense.Items, 1)
	assertIncomeExpenseItem(t, incomeExpense.Items[0], "2026-03", "0", "30")
}

func TestReportQueryService_GetBalanceReport_UserWithoutAccountsHasZeroBalance(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityDay, day(2026, 3, 1), day(2026, 3, 3)))

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2026-03-01": "0", "2026-03-02": "0", "2026-03-03": "0"}, balanceSeries(result))
}

func TestReportQueryService_GetBalanceReport_RequiresNamedLocation(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	tests := []struct {
		name     string
		location time.Location
	}{
		{"zero location", time.Location{}},
		{"local location has no name Postgres knows", *time.Local},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criteria := balanceCriteria(uuid.New(), query.GranularityDay, day(2026, 3, 1), day(2026, 3, 2))
			criteria.Location = tt.location

			result, err := qs.GetBalanceReport(ctx, criteria)

			require.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestReportQueryService_GetBalanceReport_MonthlyBucketsCrossTheYear(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "10", day(2025, 12, 31), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "1", day(2026, 1, 1), true)

	result, err := qs.GetBalanceReport(ctx, balanceCriteria(user.ID, query.GranularityMonth, day(2025, 11, 20), day(2026, 1, 15)))

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"2025-11-30": "0", "2025-12-31": "10", "2026-01-15": "11"}, balanceSeries(result))
}

func balanceStartCriteria(userID uuid.UUID) query.GetBalanceReportStartCriteria {
	return query.GetBalanceReportStartCriteria{
		UserID:   userID,
		Currency: valueObject.CurrencyUSD,
		Location: *time.UTC,
	}
}

func TestReportQueryService_GetBalanceReportStart_IsTheFirstConfirmedTransaction(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "1", day(2026, 1, 1), false)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionIncome, "10", at(2026, 3, 5, 22, 30), true)
	createSeriesTransaction(t, ctx, account.ID, enum.TransactionExpense, "5", day(2026, 4, 1), true)

	result, err := qs.GetBalanceReportStart(ctx, balanceStartCriteria(user.ID))

	require.NoError(t, err)
	require.NotNil(t, result.Date)
	assert.Equal(t, "2026-03-05", result.Date.Format("2006-01-02"), "unconfirmed transactions do not count")

	criteria := balanceStartCriteria(user.ID)
	criteria.Location = *moscowLocation(t)
	result, err = qs.GetBalanceReportStart(ctx, criteria)

	require.NoError(t, err)
	require.NotNil(t, result.Date)
	assert.Equal(t, "2026-03-06", result.Date.Format("2006-01-02"), "22:30 UTC is already the next day in Moscow")
}

func TestReportQueryService_GetBalanceReportStart_FallsBackToTheEarliestAccount(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("100"))
	createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	result, err := qs.GetBalanceReportStart(ctx, balanceStartCriteria(user.ID))

	require.NoError(t, err)
	require.NotNil(t, result.Date)
	assert.Equal(t, time.Now().UTC().Format("2006-01-02"), result.Date.Format("2006-01-02"))
}

func TestReportQueryService_GetBalanceReportStart_WithoutAccountsThereIsNoDate(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)

	result, err := qs.GetBalanceReportStart(ctx, balanceStartCriteria(user.ID))

	require.NoError(t, err)
	assert.Nil(t, result.Date)
}

func TestReportQueryService_GetBalanceReportStart_OnlyLooksAtTheSelectedAccounts(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	counted := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	selected := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	excluded := createExcludedAccount(t, ctx, user.ID, "0")
	euro := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))
	stranger := createTestUser(t, ctx)
	strangerAccount := createTestAccount(t, ctx, stranger.ID, valueObject.CurrencyUSD, dec("0"))

	createSeriesTransaction(t, ctx, counted.ID, enum.TransactionIncome, "1", day(2026, 3, 10), true)
	createSeriesTransaction(t, ctx, selected.ID, enum.TransactionIncome, "1", day(2026, 3, 20), true)
	createSeriesTransaction(t, ctx, excluded.ID, enum.TransactionIncome, "1", day(2026, 2, 1), true)
	createSeriesTransaction(t, ctx, euro.ID, enum.TransactionIncome, "1", day(2026, 1, 1), true)
	createSeriesTransaction(t, ctx, strangerAccount.ID, enum.TransactionIncome, "1", day(2025, 1, 1), true)

	criteria := balanceStartCriteria(user.ID)

	result, err := qs.GetBalanceReportStart(ctx, criteria)
	require.NoError(t, err)
	require.NotNil(t, result.Date)
	assert.Equal(t, "2026-03-10", result.Date.Format("2006-01-02"), "excluded, foreign currency and foreign users are skipped")

	criteria.AccountIDs = []uuid.UUID{selected.ID}
	result, err = qs.GetBalanceReportStart(ctx, criteria)
	require.NoError(t, err)
	require.NotNil(t, result.Date)
	assert.Equal(t, "2026-03-20", result.Date.Format("2006-01-02"))

	criteria.AccountIDs = []uuid.UUID{excluded.ID}
	result, err = qs.GetBalanceReportStart(ctx, criteria)
	require.NoError(t, err)
	require.NotNil(t, result.Date)
	assert.Equal(t, "2026-02-01", result.Date.Format("2006-01-02"), "an excluded account counts when it is selected")
}

func TestReportQueryService_GetBalanceReportStart_RequiresNamedLocation(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	criteria := balanceStartCriteria(uuid.New())
	criteria.Location = time.Location{}

	result, err := qs.GetBalanceReportStart(ctx, criteria)

	require.Error(t, err)
	assert.Nil(t, result)
}
