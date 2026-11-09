package query

import (
	"context"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetCategoryReportCriteria struct {
	UserID                 uuid.UUID
	TransactionType        enum.TransactionType
	Currency               valueObject.CurrencyCode
	PeriodStart            *time.Time
	PeriodEnd              *time.Time
	TransactionCategoryIDs []uuid.UUID
	GroupSubcategories     bool
	AccountIDs             []uuid.UUID
}

type CategoryReportItem struct {
	TransactionCategory *model.TransactionCategory
	Amount              decimal.Decimal
}

type GetCategoryReportResult struct {
	Items []CategoryReportItem
}

type GetCounterpartyReportCriteria struct {
	UserID          uuid.UUID
	TransactionType enum.TransactionType
	Currency        valueObject.CurrencyCode
	PeriodStart     *time.Time
	PeriodEnd       *time.Time
	CounterpartyIDs []uuid.UUID
	AccountIDs      []uuid.UUID
}

type CounterpartyReportItem struct {
	Counterparty *model.Counterparty
	Amount       decimal.Decimal
}

type GetCounterpartyReportResult struct {
	Items []CounterpartyReportItem
}

type GetIncomeExpenseReportCriteria struct {
	UserID      uuid.UUID
	Currency    valueObject.CurrencyCode
	Location    time.Location
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	AccountIDs  []uuid.UUID
}

type IncomeExpenseReportItem struct {
	Month   string // YYYY-MM
	Income  decimal.Decimal
	Expense decimal.Decimal
}

type GetIncomeExpenseReportResult struct {
	Items []IncomeExpenseReportItem
}

type GetBalanceReportCriteria struct {
	UserID      uuid.UUID
	Currency    valueObject.CurrencyCode
	Location    time.Location
	Granularity Granularity
	PeriodStart time.Time
	PeriodEnd   time.Time
	AccountIDs  []uuid.UUID
}

type BalanceReportItem struct {
	DateKey string
	Balance decimal.Decimal
}

type GetBalanceReportResult struct {
	Items []BalanceReportItem
}

type GetBalanceReportStartCriteria struct {
	UserID     uuid.UUID
	Currency   valueObject.CurrencyCode
	Location   time.Location
	AccountIDs []uuid.UUID
}

type GetBalanceReportStartResult struct {
	Date *time.Time
}

type GetCategoryDynamicReportCriteria struct {
	UserID                 uuid.UUID
	Currency               valueObject.CurrencyCode
	PeriodStart            *time.Time
	PeriodEnd              *time.Time
	Location               time.Location
	TransactionCategoryIDs []uuid.UUID
	AccountIDs             []uuid.UUID
}

type CategoryDynamicReportCategory struct {
	TransactionCategory model.TransactionCategory
	Amount              decimal.Decimal
}

type CategoryDynamicReportItem struct {
	Month      string // YYYY-MM
	Categories []CategoryDynamicReportCategory
}

type GetCategoryDynamicReportResult struct {
	Items []CategoryDynamicReportItem
}

type ReportQueryService interface {
	GetCategoryReport(ctx context.Context, criteria GetCategoryReportCriteria) (*GetCategoryReportResult, error)
	GetCounterpartyReport(ctx context.Context, criteria GetCounterpartyReportCriteria) (*GetCounterpartyReportResult, error)
	GetIncomeExpenseReport(ctx context.Context, criteria GetIncomeExpenseReportCriteria) (*GetIncomeExpenseReportResult, error)
	GetBalanceReport(ctx context.Context, criteria GetBalanceReportCriteria) (*GetBalanceReportResult, error)
	GetBalanceReportStart(ctx context.Context, criteria GetBalanceReportStartCriteria) (*GetBalanceReportStartResult, error)
	GetCategoryDynamicReport(ctx context.Context, criteria GetCategoryDynamicReportCriteria) (*GetCategoryDynamicReportResult, error)
}
