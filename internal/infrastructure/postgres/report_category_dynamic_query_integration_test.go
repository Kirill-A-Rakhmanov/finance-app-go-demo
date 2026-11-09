//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func categoryDynamicCriteria(userID uuid.UUID, start, end time.Time) query.GetCategoryDynamicReportCriteria {
	return query.GetCategoryDynamicReportCriteria{
		UserID:      userID,
		Currency:    valueObject.CurrencyUSD,
		Location:    *time.UTC,
		PeriodStart: &start,
		PeriodEnd:   &end,
	}
}

func categoryDynamicAmounts(item query.CategoryDynamicReportItem) map[uuid.UUID]string {
	amounts := make(map[uuid.UUID]string, len(item.Categories))
	for _, category := range item.Categories {
		amounts[category.TransactionCategory.ID] = category.Amount.String()
	}

	return amounts
}

func TestReportQueryService_GetCategoryDynamicReport_GroupsByMonthAndCategoryAndFillsEmptyMonths(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	rent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	salary := createTestCategory(t, ctx, user.ID, enum.Income, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), day(2026, 1, 5), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("15.5"), day(2026, 1, 31), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("100"), day(2026, 1, 6), true, &rent.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("900"), day(2026, 1, 10), true, &salary.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("20"), day(2026, 3, 1), true, &food.ID)

	result, err := qs.GetCategoryDynamicReport(ctx, categoryDynamicCriteria(user.ID, day(2026, 1, 1), day(2026, 3, 31)))

	require.NoError(t, err)
	require.Len(t, result.Items, 3)

	assert.Equal(t, "2026-01", result.Items[0].Month)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "25.5", rent.ID: "100", salary.ID: "900"}, categoryDynamicAmounts(result.Items[0]))
	assert.Equal(t, salary.ID, result.Items[0].Categories[0].TransactionCategory.ID, "within a month the biggest amount comes first")

	assert.Equal(t, "2026-02", result.Items[1].Month)
	assert.NotNil(t, result.Items[1].Categories)
	assert.Empty(t, result.Items[1].Categories)

	assert.Equal(t, "2026-03", result.Items[2].Month)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "20"}, categoryDynamicAmounts(result.Items[2]))
}

func TestReportQueryService_GetCategoryDynamicReport_SkipsUncategorizedTransfersAndUnconfirmed(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("1000"))
	other := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), day(2026, 1, 5), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("7"), day(2026, 1, 5), true, nil)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), day(2026, 1, 5), false, &food.ID)
	createSeriesTransfer(t, ctx, account.ID, other.ID, "50", day(2026, 1, 6))

	result, err := qs.GetCategoryDynamicReport(ctx, categoryDynamicCriteria(user.ID, day(2026, 1, 1), day(2026, 1, 31)))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, categoryDynamicAmounts(result.Items[0]))
}

func TestReportQueryService_GetCategoryDynamicReport_CategoryAndAccountFilters(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	first := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	second := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	rent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, first.ID, enum.TransactionExpense, dec("10"), day(2026, 1, 5), true, &food.ID)
	createTestTransaction(t, ctx, first.ID, enum.TransactionExpense, dec("100"), day(2026, 1, 5), true, &rent.ID)
	createTestTransaction(t, ctx, second.ID, enum.TransactionExpense, dec("3"), day(2026, 1, 5), true, &food.ID)

	criteria := categoryDynamicCriteria(user.ID, day(2026, 1, 1), day(2026, 1, 31))
	criteria.TransactionCategoryIDs = []uuid.UUID{food.ID}
	criteria.AccountIDs = []uuid.UUID{first.ID}

	result, err := qs.GetCategoryDynamicReport(ctx, criteria)

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, categoryDynamicAmounts(result.Items[0]))
}

func TestReportQueryService_GetCategoryDynamicReport_MonthsFollowTheTimezone(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	// 22:30 UTC on Jan 31 is already Feb 1 in Moscow.
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), at(2026, 1, 31, 22, 30), true, &food.ID)

	criteria := categoryDynamicCriteria(user.ID, day(2026, 1, 1), day(2026, 2, 28))
	criteria.Location = *moscowLocation(t)

	result, err := qs.GetCategoryDynamicReport(ctx, criteria)

	require.NoError(t, err)
	require.Len(t, result.Items, 2)
	assert.Empty(t, result.Items[0].Categories)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, categoryDynamicAmounts(result.Items[1]))
}

func TestReportQueryService_GetCategoryDynamicReport_OpenPeriodAndNoData(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	empty, err := qs.GetCategoryDynamicReport(ctx, query.GetCategoryDynamicReportCriteria{
		UserID: user.ID, Currency: valueObject.CurrencyUSD, Location: *time.UTC,
	})
	require.NoError(t, err)
	assert.Empty(t, empty.Items)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), day(2024, 11, 5), true, &food.ID)

	end := day(2025, 1, 31)
	result, err := qs.GetCategoryDynamicReport(ctx, query.GetCategoryDynamicReportCriteria{
		UserID: user.ID, Currency: valueObject.CurrencyUSD, Location: *time.UTC, PeriodEnd: &end,
	})
	require.NoError(t, err)
	require.Len(t, result.Items, 3)
	assert.Equal(t, "2024-11", result.Items[0].Month)
	assert.Equal(t, "2025-01", result.Items[2].Month)
}

func TestReportQueryService_GetCategoryDynamicReport_RequiresNamedLocation(t *testing.T) {
	qs := postgres.NewReportQueryService(testClient)

	_, err := qs.GetCategoryDynamicReport(context.Background(), query.GetCategoryDynamicReportCriteria{
		UserID: uuid.New(), Currency: valueObject.CurrencyUSD, Location: time.Location{},
	})

	require.Error(t, err)
}
