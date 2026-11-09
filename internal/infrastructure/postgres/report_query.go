package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/account"
	entCounterparty "finance-app-go-demo/ent/counterparty"
	"finance-app-go-demo/ent/predicate"
	entTransaction "finance-app-go-demo/ent/transaction"
	entCategory "finance-app-go-demo/ent/transactioncategory"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/pkg/transaction"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type reportQueryService struct {
	client *ent.Client
}

func NewReportQueryService(client *ent.Client) query.ReportQueryService {
	return &reportQueryService{client: client}
}

func (qs *reportQueryService) GetCategoryReport(ctx context.Context, criteria query.GetCategoryReportCriteria) (*query.GetCategoryReportResult, error) {
	client := transaction.GetClient(ctx, qs.client)

	var rows []struct {
		CategoryID *uuid.UUID      `json:"category_id"`
		Amount     decimal.Decimal `json:"amount"`
	}

	err := client.Transaction.Query().
		Where(buildReportPredicates(
			criteria.UserID,
			criteria.TransactionType,
			criteria.Currency,
			criteria.AccountIDs,
			criteria.PeriodStart,
			criteria.PeriodEnd,
		)...).
		Modify(func(s *sql.Selector) {
			applyCategoryReportGrouping(s, criteria)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	tcIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row.CategoryID != nil {
			tcIDs = append(tcIDs, *row.CategoryID)
		}
	}

	categoryMap := make(map[uuid.UUID]*ent.TransactionCategory, len(tcIDs))
	if len(tcIDs) > 0 {
		entities, err := client.TransactionCategory.Query().
			Where(entCategory.IDIn(tcIDs...)).
			All(ctx)
		if err != nil {
			return nil, err
		}

		for _, e := range entities {
			categoryMap[e.ID] = e
		}
	}

	result := &query.GetCategoryReportResult{
		Items: make([]query.CategoryReportItem, 0, len(rows)),
	}

	uncategorized := -1

	for _, row := range rows {
		var category *ent.TransactionCategory
		if row.CategoryID != nil {
			category = categoryMap[*row.CategoryID]
		}

		if category == nil {
			if uncategorized >= 0 {
				result.Items[uncategorized].Amount = result.Items[uncategorized].Amount.Add(row.Amount)
				continue
			}

			uncategorized = len(result.Items)
			result.Items = append(result.Items, query.CategoryReportItem{Amount: row.Amount})
			continue
		}

		result.Items = append(result.Items, query.CategoryReportItem{
			TransactionCategory: toTransactionCategoryModel(category),
			Amount:              row.Amount,
		})
	}

	return result, nil
}

func buildReportAccountPredicate(
	userID uuid.UUID,
	currency valueObject.CurrencyCode,
	accountIDs []uuid.UUID,
) predicate.Transaction {
	return entTransaction.HasAccountWith(buildReportAccountPredicates(userID, currency, accountIDs)...)
}

func buildReportAccountPredicates(
	userID uuid.UUID,
	currency valueObject.CurrencyCode,
	accountIDs []uuid.UUID,
) []predicate.Account {
	accountPredicates := []predicate.Account{
		account.UserIDEQ(userID),
		account.CurrencyEQ(string(currency)),
	}

	if len(accountIDs) > 0 {
		accountPredicates = append(accountPredicates, account.IDIn(accountIDs...))
	} else {
		accountPredicates = append(accountPredicates, account.IsExcludedFromCountEQ(false))
	}

	return accountPredicates
}

func buildReportPredicates(
	userID uuid.UUID,
	transactionType enum.TransactionType,
	currency valueObject.CurrencyCode,
	accountIDs []uuid.UUID,
	periodStart *time.Time,
	periodEnd *time.Time,
) []predicate.Transaction {
	predicates := []predicate.Transaction{
		buildReportAccountPredicate(userID, currency, accountIDs),
		entTransaction.TypeEQ(entTransaction.Type(transactionType)),
		entTransaction.IsConfirmedEQ(true),
		entTransaction.TransferIDIsNil(),
	}

	if periodStart != nil {
		predicates = append(predicates, entTransaction.DateGTE(*periodStart))
	}

	if periodEnd != nil {
		predicates = append(predicates, entTransaction.DateLTE(*periodEnd))
	}

	return predicates
}

func applyCategoryReportGrouping(s *sql.Selector, criteria query.GetCategoryReportCriteria) {
	groupKey := s.C(entTransaction.FieldTransactionCategoryID)

	if criteria.GroupSubcategories {
		own := sql.Table(entCategory.Table).As("own_category")
		s.LeftJoin(own).On(groupKey, own.C(entCategory.FieldID))

		groupKey = fmt.Sprintf("coalesce(%s, %s)", own.C(entCategory.FieldParentID), own.C(entCategory.FieldID))
	}

	if len(criteria.TransactionCategoryIDs) > 0 {
		tcIDs := make([]any, len(criteria.TransactionCategoryIDs))
		for i, id := range criteria.TransactionCategoryIDs {
			tcIDs[i] = id
		}

		s.Where(sql.In(groupKey, tcIDs...))
	}

	s.Select(
		sql.As(groupKey, "category_id"),
		sql.As(sql.Sum(s.C(entTransaction.FieldAmount)), "amount"),
	).
		GroupBy(groupKey).
		OrderBy(sql.Desc("amount"), groupKey)
}

func (qs *reportQueryService) GetCounterpartyReport(ctx context.Context, criteria query.GetCounterpartyReportCriteria) (*query.GetCounterpartyReportResult, error) {
	client := transaction.GetClient(ctx, qs.client)

	var rows []struct {
		CounterpartyID *uuid.UUID      `json:"counterparty_id"`
		Amount         decimal.Decimal `json:"amount"`
	}

	err := client.Transaction.Query().
		Where(buildReportPredicates(
			criteria.UserID,
			criteria.TransactionType,
			criteria.Currency,
			criteria.AccountIDs,
			criteria.PeriodStart,
			criteria.PeriodEnd,
		)...).
		Modify(func(s *sql.Selector) {
			applyCounterpartyReportGrouping(s, criteria)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	counterpartyIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row.CounterpartyID != nil {
			counterpartyIDs = append(counterpartyIDs, *row.CounterpartyID)
		}
	}

	counterpartyMap := make(map[uuid.UUID]*ent.Counterparty, len(counterpartyIDs))
	if len(counterpartyIDs) > 0 {
		entities, err := client.Counterparty.Query().
			Where(entCounterparty.IDIn(counterpartyIDs...)).
			All(ctx)
		if err != nil {
			return nil, err
		}

		for _, e := range entities {
			counterpartyMap[e.ID] = e
		}
	}

	result := &query.GetCounterpartyReportResult{
		Items: make([]query.CounterpartyReportItem, 0, len(rows)),
	}

	withoutCounterparty := -1

	for _, row := range rows {
		var counterparty *ent.Counterparty
		if row.CounterpartyID != nil {
			counterparty = counterpartyMap[*row.CounterpartyID]
		}

		if counterparty == nil {
			if withoutCounterparty >= 0 {
				result.Items[withoutCounterparty].Amount = result.Items[withoutCounterparty].Amount.Add(row.Amount)
				continue
			}

			withoutCounterparty = len(result.Items)
			result.Items = append(result.Items, query.CounterpartyReportItem{Amount: row.Amount})
			continue
		}

		result.Items = append(result.Items, query.CounterpartyReportItem{
			Counterparty: toCounterpartyModel(counterparty),
			Amount:       row.Amount,
		})
	}

	return result, nil
}

func applyCounterpartyReportGrouping(s *sql.Selector, criteria query.GetCounterpartyReportCriteria) {
	groupKey := s.C(entTransaction.FieldCounterpartyID)

	if len(criteria.CounterpartyIDs) > 0 {
		ids := make([]any, len(criteria.CounterpartyIDs))
		for i, id := range criteria.CounterpartyIDs {
			ids[i] = id
		}

		s.Where(sql.In(groupKey, ids...))
	}

	s.Select(
		sql.As(groupKey, "counterparty_id"),
		sql.As(sql.Sum(s.C(entTransaction.FieldAmount)), "amount"),
	).
		GroupBy(groupKey).
		OrderBy(sql.Desc("amount"), groupKey)
}

const monthLayout = "2006-01"

func (qs *reportQueryService) GetIncomeExpenseReport(
	ctx context.Context,
	criteria query.GetIncomeExpenseReportCriteria,
) (*query.GetIncomeExpenseReportResult, error) {
	timezone, err := reportTimezone(criteria.Location)
	if err != nil {
		return nil, err
	}

	client := transaction.GetClient(ctx, qs.client)

	var rows []struct {
		Month  time.Time           `json:"month"`
		Type   entTransaction.Type `json:"type"`
		Amount decimal.Decimal     `json:"amount"`
	}

	err = client.Transaction.Query().
		Where(buildIncomeExpenseReportPredicates(criteria)...).
		Modify(func(s *sql.Selector) {
			applyIncomeExpenseReportGrouping(s, timezone)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	totals := make(map[string]query.IncomeExpenseReportItem, len(rows))

	for _, row := range rows {
		key := row.Month.Format(monthLayout)

		item := totals[key]
		if row.Type == entTransaction.TypeIncome {
			item.Income = row.Amount
		} else {
			item.Expense = row.Amount
		}
		totals[key] = item
	}

	var dataFirst, dataLast time.Time
	if len(rows) > 0 {
		dataFirst, dataLast = rows[0].Month, rows[len(rows)-1].Month
	}

	items := make([]query.IncomeExpenseReportItem, 0)

	first, last, ok := monthRange(criteria.PeriodStart, criteria.PeriodEnd, criteria.Location, dataFirst, dataLast)
	if ok {
		for month := first; !month.After(last); month = month.AddDate(0, 1, 0) {
			key := month.Format(monthLayout)

			item := totals[key]
			item.Month = key
			items = append(items, item)
		}
	}

	return &query.GetIncomeExpenseReportResult{
		Items: items,
	}, nil
}

func reportTimezone(location time.Location) (string, error) {
	name := location.String()

	if name == "" || name == "Local" {
		return "", fmt.Errorf("report location must be a named timezone")
	}

	return name, nil
}

func applyIncomeExpenseReportGrouping(s *sql.Selector, timezone string) {
	month := sql.P(func(b *sql.Builder) {
		b.WriteString("date_trunc('month', ").
			WriteString(s.C(entTransaction.FieldDate)).
			WriteString(" AT TIME ZONE ").
			Arg(timezone).
			WriteString(")")
	})
	transactionType := s.C(entTransaction.FieldType)

	s.Select(
		sql.As(transactionType, "type"),
		sql.As(sql.Sum(s.C(entTransaction.FieldAmount)), "amount"),
	).
		AppendSelectExprAs(month, "month").
		GroupBy("month", transactionType).
		OrderBy("month")
}

func buildIncomeExpenseReportPredicates(criteria query.GetIncomeExpenseReportCriteria) []predicate.Transaction {
	return buildMonthlyReportPredicates(
		criteria.UserID,
		criteria.Currency,
		criteria.AccountIDs,
		criteria.PeriodStart,
		criteria.PeriodEnd,
		criteria.Location,
	)
}

func buildMonthlyReportPredicates(
	userID uuid.UUID,
	currency valueObject.CurrencyCode,
	accountIDs []uuid.UUID,
	periodStart *time.Time,
	periodEnd *time.Time,
	location time.Location,
) []predicate.Transaction {
	predicates := []predicate.Transaction{
		buildReportAccountPredicate(userID, currency, accountIDs),
		entTransaction.IsConfirmedEQ(true),
		entTransaction.TransferIDIsNil(),
	}

	if periodStart != nil {
		predicates = append(predicates, entTransaction.DateGTE(startOfDay(*periodStart, location)))
	}

	if periodEnd != nil {
		nextDay := startOfDay(*periodEnd, location).AddDate(0, 0, 1)
		predicates = append(predicates, entTransaction.DateLT(nextDay))
	}

	return predicates
}

func monthRange(
	periodStart,
	periodEnd *time.Time,
	location time.Location,
	dataFirst,
	dataLast time.Time,
) (first, last time.Time, ok bool) {
	switch {
	case periodStart != nil:
		first = startOfMonth(*periodStart, location)
	case !dataFirst.IsZero():
		first = startOfMonth(dataFirst, location)
	default:
		return first, last, false
	}

	if periodEnd != nil {
		last = startOfMonth(*periodEnd, location)
	} else {
		last = startOfMonth(time.Now().In(&location), location)

		if !dataLast.IsZero() {
			if lastWithData := startOfMonth(dataLast, location); lastWithData.After(last) {
				last = lastWithData
			}
		}
	}

	return first, last, !first.After(last)
}

func (qs *reportQueryService) GetCategoryDynamicReport(
	ctx context.Context,
	criteria query.GetCategoryDynamicReportCriteria,
) (*query.GetCategoryDynamicReportResult, error) {
	timezone, err := reportTimezone(criteria.Location)
	if err != nil {
		return nil, err
	}

	client := transaction.GetClient(ctx, qs.client)

	predicates := buildMonthlyReportPredicates(
		criteria.UserID,
		criteria.Currency,
		criteria.AccountIDs,
		criteria.PeriodStart,
		criteria.PeriodEnd,
		criteria.Location,
	)
	predicates = append(predicates, entTransaction.TransactionCategoryIDNotNil())
	if len(criteria.TransactionCategoryIDs) > 0 {
		predicates = append(predicates, entTransaction.TransactionCategoryIDIn(criteria.TransactionCategoryIDs...))
	}

	var rows []struct {
		Month      time.Time       `json:"month"`
		CategoryID uuid.UUID       `json:"category_id"`
		Amount     decimal.Decimal `json:"amount"`
	}

	err = client.Transaction.Query().
		Where(predicates...).
		Modify(func(s *sql.Selector) {
			applyCategoryDynamicReportGrouping(s, timezone)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.CategoryID]; !ok {
			seen[row.CategoryID] = struct{}{}
			categoryIDs = append(categoryIDs, row.CategoryID)
		}
	}

	categoryMap := make(map[uuid.UUID]*ent.TransactionCategory, len(categoryIDs))
	if len(categoryIDs) > 0 {
		entities, err := client.TransactionCategory.Query().
			Where(entCategory.IDIn(categoryIDs...)).
			All(ctx)
		if err != nil {
			return nil, err
		}

		for _, e := range entities {
			categoryMap[e.ID] = e
		}
	}

	byMonth := make(map[string][]query.CategoryDynamicReportCategory)

	for _, row := range rows {
		category := categoryMap[row.CategoryID]
		if category == nil {
			continue
		}

		key := row.Month.Format(monthLayout)
		byMonth[key] = append(byMonth[key], query.CategoryDynamicReportCategory{
			TransactionCategory: *toTransactionCategoryModel(category),
			Amount:              row.Amount,
		})
	}

	var dataFirst, dataLast time.Time
	if len(rows) > 0 {
		dataFirst, dataLast = rows[0].Month, rows[len(rows)-1].Month
	}

	items := make([]query.CategoryDynamicReportItem, 0)

	first, last, ok := monthRange(criteria.PeriodStart, criteria.PeriodEnd, criteria.Location, dataFirst, dataLast)
	if ok {
		for month := first; !month.After(last); month = month.AddDate(0, 1, 0) {
			key := month.Format(monthLayout)

			categories := byMonth[key]
			if categories == nil {
				categories = []query.CategoryDynamicReportCategory{}
			}

			items = append(items, query.CategoryDynamicReportItem{Month: key, Categories: categories})
		}
	}

	return &query.GetCategoryDynamicReportResult{Items: items}, nil
}

func applyCategoryDynamicReportGrouping(s *sql.Selector, timezone string) {
	month := sql.P(func(b *sql.Builder) {
		b.WriteString("date_trunc('month', ").
			WriteString(s.C(entTransaction.FieldDate)).
			WriteString(" AT TIME ZONE ").
			Arg(timezone).
			WriteString(")")
	})
	categoryID := s.C(entTransaction.FieldTransactionCategoryID)

	s.Select(
		sql.As(categoryID, "category_id"),
		sql.As(sql.Sum(s.C(entTransaction.FieldAmount)), "amount"),
	).
		AppendSelectExprAs(month, "month").
		GroupBy("month", categoryID).
		OrderBy("month", sql.Desc("amount"), categoryID)
}

func (qs *reportQueryService) GetBalanceReport(
	ctx context.Context,
	criteria query.GetBalanceReportCriteria,
) (*query.GetBalanceReportResult, error) {
	timezone, err := reportTimezone(criteria.Location)
	if err != nil {
		return nil, err
	}

	unit, err := granularityUnit(criteria.Granularity)
	if err != nil {
		return nil, err
	}

	client := transaction.GetClient(ctx, qs.client)

	periodStart := startOfDay(criteria.PeriodStart, criteria.Location)
	periodEnd := startOfDay(criteria.PeriodEnd, criteria.Location)

	balance, err := getBalanceBefore(ctx, client, criteria, periodStart)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		Bucket time.Time           `json:"bucket"`
		Type   entTransaction.Type `json:"type"`
		Amount decimal.Decimal     `json:"amount"`
	}

	err = client.Transaction.Query().
		Where(buildBalanceReportPredicates(criteria, periodStart, periodEnd)...).
		Modify(func(s *sql.Selector) {
			applyBalanceReportGrouping(s, timezone, unit)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	changes := make(map[string]decimal.Decimal, len(rows))

	for _, row := range rows {
		key := row.Bucket.Format(dateLayout)
		changes[key] = changes[key].Add(signedAmount(row.Type, row.Amount))
	}

	items := make([]query.BalanceReportItem, 0)

	for _, bucket := range buildBalanceBuckets(periodStart, periodEnd, criteria.Granularity) {
		balance = balance.Add(changes[bucket.start.Format(dateLayout)])

		items = append(items, query.BalanceReportItem{
			DateKey: bucket.end.Format(dateLayout),
			Balance: balance,
		})
	}

	return &query.GetBalanceReportResult{
		Items: items,
	}, nil
}

func (qs *reportQueryService) GetBalanceReportStart(
	ctx context.Context,
	criteria query.GetBalanceReportStartCriteria,
) (*query.GetBalanceReportStartResult, error) {
	if _, err := reportTimezone(criteria.Location); err != nil {
		return nil, err
	}

	client := transaction.GetClient(ctx, qs.client)

	first, err := client.Transaction.Query().
		Where(
			buildReportAccountPredicate(criteria.UserID, criteria.Currency, criteria.AccountIDs),
			entTransaction.IsConfirmedEQ(true),
		).
		Order(entTransaction.ByDate()).
		First(ctx)
	if err == nil {
		return balanceReportStart(first.Date, criteria.Location), nil
	}

	if !ent.IsNotFound(err) {
		return nil, err
	}

	earliestAccount, err := client.Account.Query().
		Where(buildReportAccountPredicates(criteria.UserID, criteria.Currency, criteria.AccountIDs)...).
		Order(account.ByCreatedAt()).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &query.GetBalanceReportStartResult{}, nil
		}

		return nil, err
	}

	return balanceReportStart(earliestAccount.CreatedAt, criteria.Location), nil
}

func balanceReportStart(moment time.Time, location time.Location) *query.GetBalanceReportStartResult {
	date := startOfDay(moment.In(&location), location)

	return &query.GetBalanceReportStartResult{Date: &date}
}

func getBalanceBefore(
	ctx context.Context,
	client *ent.Client,
	criteria query.GetBalanceReportCriteria,
	periodStart time.Time,
) (decimal.Decimal, error) {
	accounts, err := client.Account.Query().
		Where(buildReportAccountPredicates(criteria.UserID, criteria.Currency, criteria.AccountIDs)...).
		All(ctx)
	if err != nil {
		return decimal.Zero, err
	}

	balance := decimal.Zero
	for _, acc := range accounts {
		balance = balance.Add(acc.Adjustment)
	}

	var rows []struct {
		Type   entTransaction.Type `json:"type"`
		Amount decimal.Decimal     `json:"amount"`
	}

	err = client.Transaction.Query().
		Where(
			buildReportAccountPredicate(criteria.UserID, criteria.Currency, criteria.AccountIDs),
			entTransaction.IsConfirmedEQ(true),
			entTransaction.DateLT(periodStart),
		).
		GroupBy(entTransaction.FieldType).
		Aggregate(ent.As(ent.Sum(entTransaction.FieldAmount), "amount")).
		Scan(ctx, &rows)
	if err != nil {
		return decimal.Zero, err
	}

	for _, row := range rows {
		balance = balance.Add(signedAmount(row.Type, row.Amount))
	}

	return balance, nil
}

func signedAmount(transactionType entTransaction.Type, amount decimal.Decimal) decimal.Decimal {
	if transactionType == entTransaction.TypeIncome {
		return amount
	}

	return amount.Neg()
}

func buildBalanceReportPredicates(
	criteria query.GetBalanceReportCriteria,
	periodStart,
	periodEnd time.Time,
) []predicate.Transaction {
	return []predicate.Transaction{
		buildReportAccountPredicate(criteria.UserID, criteria.Currency, criteria.AccountIDs),
		entTransaction.IsConfirmedEQ(true),
		entTransaction.DateGTE(periodStart),
		entTransaction.DateLT(periodEnd.AddDate(0, 0, 1)),
	}
}

func applyBalanceReportGrouping(s *sql.Selector, timezone, unit string) {
	bucket := sql.P(func(b *sql.Builder) {
		b.WriteString("date_trunc(").
			Arg(unit).
			WriteString(", ").
			WriteString(s.C(entTransaction.FieldDate)).
			WriteString(" AT TIME ZONE ").
			Arg(timezone).
			WriteString(")")
	})
	transactionType := s.C(entTransaction.FieldType)

	s.Select(
		sql.As(transactionType, "type"),
		sql.As(sql.Sum(s.C(entTransaction.FieldAmount)), "amount"),
	).
		AppendSelectExprAs(bucket, "bucket").
		GroupBy("bucket", transactionType)
}

type balanceBucket struct {
	start time.Time
	end   time.Time
}

func buildBalanceBuckets(periodStart, periodEnd time.Time, granularity query.Granularity) []balanceBucket {
	buckets := make([]balanceBucket, 0)

	for start := startOfBucket(periodStart, granularity); !start.After(periodEnd); {
		next := nextBucketStart(start, granularity)

		end := next.AddDate(0, 0, -1)
		if end.After(periodEnd) {
			end = periodEnd
		}

		buckets = append(buckets, balanceBucket{start: start, end: end})
		start = next
	}

	return buckets
}

func startOfBucket(day time.Time, granularity query.Granularity) time.Time {
	switch granularity {
	case query.GranularityWeek:
		daysSinceMonday := (int(day.Weekday()) + 6) % 7
		return day.AddDate(0, 0, -daysSinceMonday)
	case query.GranularityMonth:
		return time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
	default:
		return day
	}
}

func nextBucketStart(start time.Time, granularity query.Granularity) time.Time {
	switch granularity {
	case query.GranularityWeek:
		return start.AddDate(0, 0, 7)
	case query.GranularityMonth:
		return start.AddDate(0, 1, 0)
	default:
		return start.AddDate(0, 0, 1)
	}
}

func granularityUnit(granularity query.Granularity) (string, error) {
	switch granularity {
	case query.GranularityDay, query.GranularityWeek, query.GranularityMonth:
		return string(granularity), nil
	default:
		return "", fmt.Errorf("unsupported report granularity %q", granularity)
	}
}

const dateLayout = "2006-01-02"

func startOfDay(t time.Time, location time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, &location)
}

func startOfMonth(t time.Time, location time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, &location)
}
