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
	"github.com/stretchr/testify/require"
)

type queryFixture struct {
	user          *model.User
	accountA      *model.Account
	accountB      *model.Account
	food          *model.TransactionCategory
	salary        *model.TransactionCategory
	oldest        *model.Transaction // accountA, food,   amount 30, 10 days ago
	middle        *model.Transaction // accountA, salary, amount 10, 5 days ago
	newest        *model.Transaction // accountB, food,   amount 20, 1 day ago
	uncategorized *model.Transaction // accountB, none,  amount 40, 3 days ago
}

func newQueryFixture(t *testing.T, ctx context.Context) queryFixture {
	t.Helper()

	user := createTestUser(t, ctx)
	f := queryFixture{
		user:     user,
		accountA: createTestAccount(t, ctx, user.ID, valueObject.CurrencyUSD, dec("0")),
		accountB: createTestAccount(t, ctx, user.ID, valueObject.CurrencyEUR, dec("0")),
		food:     createTestCategory(t, ctx, user.ID, enum.Expense, nil),
		salary:   createTestCategory(t, ctx, user.ID, enum.Income, nil),
	}

	f.oldest = createTestTransaction(t, ctx, f.accountA.ID, enum.TransactionExpense, dec("30"), testDate(10), true, &f.food.ID)
	f.middle = createTestTransaction(t, ctx, f.accountA.ID, enum.TransactionIncome, dec("10"), testDate(5), true, &f.salary.ID)
	f.newest = createTestTransaction(t, ctx, f.accountB.ID, enum.TransactionExpense, dec("20"), testDate(1), true, &f.food.ID)
	f.uncategorized = createTestTransaction(t, ctx, f.accountB.ID, enum.TransactionExpense, dec("40"), testDate(3), true, nil)

	return f
}

func ids(txs []model.Transaction) []uuid.UUID {
	result := make([]uuid.UUID, len(txs))
	for i, tx := range txs {
		result[i] = tx.ID
	}

	return result
}

func sortBy(by query.TransactionSortBy, dir query.SortDirection) (*query.TransactionSortBy, *query.SortDirection) {
	return &by, &dir
}

func TestTransactionQueryService_NoFilters_ReturnsAllOfUser(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	// another user's data must never leak in
	other := createTestUser(t, ctx)
	otherAccount := createTestAccount(t, ctx, other.ID, valueObject.CurrencyUSD, dec("0"))
	createTestTransaction(t, ctx, otherAccount.ID, enum.TransactionIncome, dec("1"), testDate(0), true, nil)

	res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{UserID: f.user.ID, Limit: 50})
	require.NoError(t, err)
	require.Equal(t, 4, res.Total)
	require.Len(t, res.Transactions, 4)
	require.Equal(t, ids(res.Transactions), res.Order)
}

func TestTransactionQueryService_DefaultOrder_NewestCreatedFirst(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{UserID: f.user.ID, Limit: 50})
	require.NoError(t, err)

	// created_at DESC: the last one inserted comes first
	require.Equal(t, []uuid.UUID{f.uncategorized.ID, f.newest.ID, f.middle.ID, f.oldest.ID}, res.Order)
}

func TestTransactionQueryService_NoTransactions(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	user := createTestUser(t, ctx)

	res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{UserID: user.ID, Limit: 10})
	require.NoError(t, err)
	require.Zero(t, res.Total)
	require.NotNil(t, res.Order)
	require.Empty(t, res.Order)
	require.NotNil(t, res.Transactions)
	require.Empty(t, res.Transactions)
}

func TestTransactionQueryService_FilterByAccount(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{UserID: f.user.ID, AccountID: &f.accountA.ID, Limit: 50})
	require.NoError(t, err)
	require.Equal(t, 2, res.Total)
	require.ElementsMatch(t, []uuid.UUID{f.oldest.ID, f.middle.ID}, res.Order)
}

func TestTransactionQueryService_FilterByAccount_OfAnotherUser(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)
	other := createTestUser(t, ctx)

	res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{UserID: other.ID, AccountID: &f.accountA.ID, Limit: 50})
	require.NoError(t, err)
	require.Zero(t, res.Total, "a foreign account id must yield nothing")
}

func TestTransactionQueryService_FilterByCategories(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	t.Run("single category", func(t *testing.T) {
		res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
			UserID: f.user.ID, TransactionCategoryIDs: []uuid.UUID{f.food.ID}, Limit: 50,
		})
		require.NoError(t, err)
		require.ElementsMatch(t, []uuid.UUID{f.oldest.ID, f.newest.ID}, res.Order)
	})

	t.Run("several categories", func(t *testing.T) {
		res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
			UserID: f.user.ID, TransactionCategoryIDs: []uuid.UUID{f.food.ID, f.salary.ID}, Limit: 50,
		})
		require.NoError(t, err)
		require.Equal(t, 3, res.Total)
		require.ElementsMatch(t, []uuid.UUID{f.oldest.ID, f.middle.ID, f.newest.ID}, res.Order)
	})

	t.Run("unknown category", func(t *testing.T) {
		res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
			UserID: f.user.ID, TransactionCategoryIDs: []uuid.UUID{uuid.New()}, Limit: 50,
		})
		require.NoError(t, err)
		require.Zero(t, res.Total)
	})
}

func TestTransactionQueryService_FilterByPeriod(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	tests := []struct {
		name  string
		start *time.Time
		end   *time.Time
		want  []uuid.UUID
	}{
		{
			name:  "start only",
			start: ptr(testDate(5)),
			want:  []uuid.UUID{f.middle.ID, f.uncategorized.ID, f.newest.ID},
		},
		{
			name: "end only",
			end:  ptr(testDate(5)),
			want: []uuid.UUID{f.oldest.ID, f.middle.ID},
		},
		{
			name:  "both bounds are inclusive",
			start: ptr(testDate(5)),
			end:   ptr(testDate(3)),
			want:  []uuid.UUID{f.middle.ID, f.uncategorized.ID},
		},
		{
			name:  "empty window",
			start: ptr(testDate(9)),
			end:   ptr(testDate(6)),
			want:  []uuid.UUID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
				UserID: f.user.ID, PeriodStart: tt.start, PeriodEnd: tt.end, Limit: 50,
			})
			require.NoError(t, err)
			require.Equal(t, len(tt.want), res.Total)
			require.ElementsMatch(t, tt.want, res.Order)
		})
	}
}

func TestTransactionQueryService_CombinedFilters(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
		UserID:                 f.user.ID,
		AccountID:              &f.accountB.ID,
		TransactionCategoryIDs: []uuid.UUID{f.food.ID},
		PeriodStart:            ptr(testDate(4)),
		Limit:                  50,
	})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{f.newest.ID}, res.Order)
}

func TestTransactionQueryService_Sorting(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	tests := []struct {
		name string
		by   query.TransactionSortBy
		dir  query.SortDirection
		want []uuid.UUID
	}{
		{"date asc", query.TransactionSortByDate, query.Ascending, []uuid.UUID{f.oldest.ID, f.middle.ID, f.uncategorized.ID, f.newest.ID}},
		{"date desc", query.TransactionSortByDate, query.Descending, []uuid.UUID{f.newest.ID, f.uncategorized.ID, f.middle.ID, f.oldest.ID}},
		{"amount asc", query.TransactionSortByAmount, query.Ascending, []uuid.UUID{f.middle.ID, f.newest.ID, f.oldest.ID, f.uncategorized.ID}},
		{"amount desc", query.TransactionSortByAmount, query.Descending, []uuid.UUID{f.uncategorized.ID, f.oldest.ID, f.newest.ID, f.middle.ID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			by, dir := sortBy(tt.by, tt.dir)
			res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
				UserID: f.user.ID, SortBy: by, SortDirection: dir, Limit: 50,
			})
			require.NoError(t, err)
			require.Equal(t, tt.want, res.Order)
			require.Equal(t, tt.want, ids(res.Transactions))
		})
	}
}

func TestTransactionQueryService_Pagination(t *testing.T) {
	ctx := context.Background()
	qs := postgres.NewTransactionQueryService(testClient)
	f := newQueryFixture(t, ctx)

	by, dir := sortBy(query.TransactionSortByDate, query.Ascending)
	all := []uuid.UUID{f.oldest.ID, f.middle.ID, f.uncategorized.ID, f.newest.ID}

	tests := []struct {
		name          string
		limit, offset int
		want          []uuid.UUID
	}{
		{"first page", 2, 0, all[:2]},
		{"second page", 2, 2, all[2:]},
		{"partial last page", 3, 3, all[3:]},
		{"offset past the end", 2, 10, []uuid.UUID{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := qs.GetTransactions(ctx, query.GetTransactionCriteria{
				UserID: f.user.ID, SortBy: by, SortDirection: dir, Limit: tt.limit, Offset: tt.offset,
			})
			require.NoError(t, err)
			require.Equal(t, 4, res.Total, "Total ignores pagination")
			require.Equal(t, tt.want, res.Order)
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
