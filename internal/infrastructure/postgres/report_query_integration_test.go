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

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReportCriteria(userID uuid.UUID) query.GetCategoryReportCriteria {
	return query.GetCategoryReportCriteria{
		UserID:          userID,
		TransactionType: enum.TransactionExpense,
		Currency:        valueObject.CurrencyUSD,
	}
}

func reportAmounts(result *query.GetCategoryReportResult) map[uuid.UUID]string {
	amounts := make(map[uuid.UUID]string, len(result.Items))
	for _, item := range result.Items {
		if item.TransactionCategory == nil {
			continue
		}

		amounts[item.TransactionCategory.ID] = item.Amount.String()
	}

	return amounts
}

func uncategorizedAmount(t *testing.T, result *query.GetCategoryReportResult) decimal.Decimal {
	t.Helper()

	amount := decimal.Zero
	found := 0

	for _, item := range result.Items {
		if item.TransactionCategory == nil {
			amount = item.Amount
			found++
		}
	}

	require.LessOrEqual(t, found, 1, "all transactions without a category must be one item")

	return amount
}

func TestReportQueryService_GetCategoryReport_GroupsAndSortsByAmount(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	rent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	fun := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(3), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("15.5"), testDate(2), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("100"), testDate(2), true, &rent.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("20"), testDate(1), true, &fun.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("7"), testDate(1), true, nil)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("3"), testDate(0), true, nil)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 4)
	assert.Equal(t, rent.ID, result.Items[0].TransactionCategory.ID)
	assert.True(t, result.Items[0].Amount.Equal(dec("100")), "got %s", result.Items[0].Amount)
	assert.Equal(t, food.ID, result.Items[1].TransactionCategory.ID)
	assert.True(t, result.Items[1].Amount.Equal(dec("25.5")), "got %s", result.Items[1].Amount)
	assert.Equal(t, fun.ID, result.Items[2].TransactionCategory.ID)
	assert.True(t, result.Items[2].Amount.Equal(dec("20")), "got %s", result.Items[2].Amount)
	assert.Nil(t, result.Items[3].TransactionCategory, "transactions without a category come as the last, smallest item")
	assert.True(t, result.Items[3].Amount.Equal(dec("10")), "got %s", result.Items[3].Amount)

	assert.Equal(t, user.ID, result.Items[0].TransactionCategory.UserID)
	assert.Equal(t, rent.Name, result.Items[0].TransactionCategory.Name)
	assert.Equal(t, enum.Expense, result.Items[0].TransactionCategory.Type)
}

func TestReportQueryService_GetCategoryReport_TieBreaksByCategoryID(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	a := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	b := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), testDate(0), true, &a.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), testDate(0), true, &b.ID)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 2)

	first, second := a.ID, b.ID
	if second.String() < first.String() {
		first, second = second, first
	}
	assert.Equal(t, first, result.Items[0].TransactionCategory.ID)
	assert.Equal(t, second, result.Items[1].TransactionCategory.ID)
}

func TestReportQueryService_GetCategoryReport_EmptyReport(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	assert.NotNil(t, result.Items)
	assert.Empty(t, result.Items)
	assert.True(t, uncategorizedAmount(t, result).IsZero())
}

func TestReportQueryService_GetCategoryReport_OnlyUncategorized(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("12"), testDate(0), true, nil)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	assert.Nil(t, result.Items[0].TransactionCategory)
	assert.True(t, result.Items[0].Amount.Equal(dec("12")))
}

func TestReportQueryService_GetCategoryReport_UncategorizedIsSortedByAmount(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	rent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("50"), testDate(0), true, nil)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("100"), testDate(0), true, &rent.ID)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 3)
	assert.Equal(t, rent.ID, result.Items[0].TransactionCategory.ID)
	assert.Nil(t, result.Items[1].TransactionCategory, "the item without a category sits between the others by amount")
	assert.True(t, result.Items[1].Amount.Equal(dec("50")), "got %s", result.Items[1].Amount)
	assert.Equal(t, food.ID, result.Items[2].TransactionCategory.ID)
}

func TestReportQueryService_GetCategoryReport_TypeFilter(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	salary := createTestCategory(t, ctx, user.ID, enum.Income, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionIncome, dec("500"), testDate(0), true, &salary.ID)

	criteria := newReportCriteria(user.ID)
	criteria.TransactionType = enum.TransactionIncome

	result, err := qs.GetCategoryReport(ctx, criteria)

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{salary.ID: "500"}, reportAmounts(result))
}

func TestReportQueryService_GetCategoryReport_OnlyConfirmedTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("99"), testDate(0), false, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("77"), testDate(0), false, nil)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, reportAmounts(result))
	assert.True(t, uncategorizedAmount(t, result).IsZero())
}

func TestReportQueryService_GetCategoryReport_ExcludesTransferTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	from := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	to := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, from.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &food.ID)

	transfer, err := postgres.NewTransferRepository(testClient).Create(ctx, model.NewTransfer(from.ID, to.ID, dec("300"), dec("300"), testDate(0)))
	require.NoError(t, err)
	_, err = postgres.NewTransactionRepository(testClient).Create(ctx, model.NewTransaction(
		from.ID, nil, nil, &transfer.ID, enum.TransactionExpense, dec("300"), nil, testDate(0), true,
	))
	require.NoError(t, err)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, reportAmounts(result))
	assert.True(t, uncategorizedAmount(t, result).IsZero(), "transfer leg must not land in the uncategorized bucket")
}

func TestReportQueryService_GetCategoryReport_CurrencySeparation(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	usd := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	eur := createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, usd.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, eur.ID, enum.TransactionExpense, dec("400"), testDate(0), true, &food.ID)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, reportAmounts(result))

	criteria := newReportCriteria(user.ID)
	criteria.Currency = valueObject.CurrencyEUR
	result, err = qs.GetCategoryReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "400"}, reportAmounts(result))
}

func TestReportQueryService_GetCategoryReport_IsolatesUsers(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	otherAccount := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	otherFood := createTestCategory(t, ctx, other.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("10"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, otherAccount.ID, enum.TransactionExpense, dec("999"), testDate(0), true, &otherFood.ID)
	createTestTransaction(t, ctx, otherAccount.ID, enum.TransactionExpense, dec("888"), testDate(0), true, nil)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "10"}, reportAmounts(result))
	assert.True(t, uncategorizedAmount(t, result).IsZero())
}

func TestReportQueryService_GetCategoryReport_PeriodBoundsAreInclusive(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(10), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("2"), testDate(5), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("4"), testDate(3), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("8"), testDate(1), true, &food.ID)

	start, end := testDate(5), testDate(3)
	criteria := newReportCriteria(user.ID)
	criteria.PeriodStart = &start
	criteria.PeriodEnd = &end

	result, err := qs.GetCategoryReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "6"}, reportAmounts(result))

	criteria = newReportCriteria(user.ID)
	criteria.PeriodStart = &start

	result, err = qs.GetCategoryReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "14"}, reportAmounts(result))

	criteria = newReportCriteria(user.ID)
	criteria.PeriodEnd = &end

	result, err = qs.GetCategoryReport(ctx, criteria)
	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{food.ID: "7"}, reportAmounts(result))
}

func TestReportQueryService_GetCategoryReport_AccountsFilter(t *testing.T) {
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
	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)

	createTestTransaction(t, ctx, regular.ID, enum.TransactionExpense, dec("1"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, another.ID, enum.TransactionExpense, dec("2"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, excluded.ID, enum.TransactionExpense, dec("4"), testDate(0), true, &food.ID)

	t.Run("no accounts means all accounts except excluded ones", func(t *testing.T) {
		result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{food.ID: "3"}, reportAmounts(result))
	})

	t.Run("explicit accounts restrict the report", func(t *testing.T) {
		criteria := newReportCriteria(user.ID)
		criteria.AccountIDs = []uuid.UUID{another.ID}

		result, err := qs.GetCategoryReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{food.ID: "2"}, reportAmounts(result))
	})

	t.Run("explicitly requested excluded account is included", func(t *testing.T) {
		criteria := newReportCriteria(user.ID)
		criteria.AccountIDs = []uuid.UUID{regular.ID, excluded.ID}

		result, err := qs.GetCategoryReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{food.ID: "5"}, reportAmounts(result))
	})
}

type hierarchyFixture struct {
	user    *model.User
	account *model.Account
	food    *model.TransactionCategory // parent, own transactions: 100
	cafe    *model.TransactionCategory // child of food: 20 + 5
	market  *model.TransactionCategory // child of food: 30
	rent    *model.TransactionCategory // no children: 400
	fun     *model.TransactionCategory // parent without own transactions, child: 7
	cinema  *model.TransactionCategory // child of fun: 7
}

func newHierarchyFixture(t *testing.T, ctx context.Context) hierarchyFixture {
	t.Helper()

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))

	food := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	cafe := createTestCategory(t, ctx, user.ID, enum.Expense, &food.ID)
	market := createTestCategory(t, ctx, user.ID, enum.Expense, &food.ID)
	rent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	fun := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	cinema := createTestCategory(t, ctx, user.ID, enum.Expense, &fun.ID)

	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("100"), testDate(0), true, &food.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("20"), testDate(0), true, &cafe.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("5"), testDate(0), true, &cafe.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("30"), testDate(0), true, &market.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("400"), testDate(0), true, &rent.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("7"), testDate(0), true, &cinema.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("3"), testDate(0), true, nil)

	return hierarchyFixture{user: user, account: account, food: food, cafe: cafe, market: market, rent: rent, fun: fun, cinema: cinema}
}

func TestReportQueryService_GetCategoryReport_WithoutGroupingKeepsSubcategoriesSeparate(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)
	f := newHierarchyFixture(t, ctx)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(f.user.ID))

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{
		f.food.ID:   "100",
		f.cafe.ID:   "25",
		f.market.ID: "30",
		f.rent.ID:   "400",
		f.cinema.ID: "7",
	}, reportAmounts(result))
	assert.True(t, uncategorizedAmount(t, result).Equal(dec("3")))
	assert.Equal(t, f.rent.ID, result.Items[0].TransactionCategory.ID)
}

func TestReportQueryService_GetCategoryReport_GroupingRollsUpIntoParents(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)
	f := newHierarchyFixture(t, ctx)

	criteria := newReportCriteria(f.user.ID)
	criteria.GroupSubcategories = true

	result, err := qs.GetCategoryReport(ctx, criteria)

	require.NoError(t, err)
	assert.Equal(t, map[uuid.UUID]string{
		f.food.ID: "155",
		f.rent.ID: "400",
		f.fun.ID:  "7",
	}, reportAmounts(result), "a parent without own transactions still appears through its children")
	assert.True(t, uncategorizedAmount(t, result).Equal(dec("3")))

	require.Len(t, result.Items, 4)
	assert.Equal(t, f.rent.ID, result.Items[0].TransactionCategory.ID)
	assert.Equal(t, f.food.ID, result.Items[1].TransactionCategory.ID)
	assert.Equal(t, f.fun.ID, result.Items[2].TransactionCategory.ID)
	assert.Nil(t, result.Items[3].TransactionCategory)
	assert.Nil(t, result.Items[1].TransactionCategory.ParentID)
}

func TestReportQueryService_GetCategoryReport_CategoryFilter(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)
	f := newHierarchyFixture(t, ctx)

	t.Run("grouped filter by parent includes its children", func(t *testing.T) {
		criteria := newReportCriteria(f.user.ID)
		criteria.GroupSubcategories = true
		criteria.TransactionCategoryIDs = []uuid.UUID{f.food.ID}

		result, err := qs.GetCategoryReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{f.food.ID: "155"}, reportAmounts(result))
		assert.True(t, uncategorizedAmount(t, result).IsZero(), "uncategorized is outside a category filter")
	})

	t.Run("not grouped filter by parent ignores its children", func(t *testing.T) {
		criteria := newReportCriteria(f.user.ID)
		criteria.TransactionCategoryIDs = []uuid.UUID{f.food.ID}

		result, err := qs.GetCategoryReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{f.food.ID: "100"}, reportAmounts(result))
	})

	t.Run("not grouped filter by subcategory", func(t *testing.T) {
		criteria := newReportCriteria(f.user.ID)
		criteria.TransactionCategoryIDs = []uuid.UUID{f.cafe.ID, f.rent.ID}

		result, err := qs.GetCategoryReport(ctx, criteria)

		require.NoError(t, err)
		assert.Equal(t, map[uuid.UUID]string{f.cafe.ID: "25", f.rent.ID: "400"}, reportAmounts(result))
	})

	t.Run("filter of another type yields an empty report", func(t *testing.T) {
		criteria := newReportCriteria(f.user.ID)
		criteria.TransactionType = enum.TransactionIncome
		criteria.TransactionCategoryIDs = []uuid.UUID{f.food.ID}

		result, err := qs.GetCategoryReport(ctx, criteria)

		require.NoError(t, err)
		assert.Empty(t, result.Items)
	})
}

func TestReportQueryService_GetCategoryReport_MapsCategoryFields(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewReportQueryService(testClient)

	user := createTestUser(t, ctx)
	account := createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0"))
	parent := createTestCategory(t, ctx, user.ID, enum.Expense, nil)
	child := createTestCategory(t, ctx, user.ID, enum.Expense, &parent.ID)
	createTestTransaction(t, ctx, account.ID, enum.TransactionExpense, dec("1"), testDate(0), true, &child.ID)

	result, err := qs.GetCategoryReport(ctx, newReportCriteria(user.ID))

	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	got := result.Items[0].TransactionCategory
	assert.Equal(t, child.ID, got.ID)
	assert.Equal(t, user.ID, got.UserID)
	require.NotNil(t, got.ParentID)
	assert.Equal(t, parent.ID, *got.ParentID)
	assert.Equal(t, child.Name, got.Name)
	assert.Equal(t, child.Icon, got.Icon)
	assert.Equal(t, child.Color, got.Color)
	assert.Equal(t, child.IsDefault, got.IsDefault)
	assert.Equal(t, child.Slug, got.Slug)
	assert.Equal(t, child.Type, got.Type)
	assert.False(t, got.CreatedAt.IsZero())
}
