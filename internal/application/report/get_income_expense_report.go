package report

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/valueObject"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetIncomeExpenseReportInput struct {
	UserID      uuid.UUID
	Currency    string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	AccountIDs  []uuid.UUID
}

type IncomeExpenseReportItem struct {
	Month   string
	Income  decimal.Decimal
	Expense decimal.Decimal
}

type GetIncomeExpenseReportResult struct {
	Items []IncomeExpenseReportItem
}

func (s *Service) GetIncomeExpenseReport(ctx context.Context, input GetIncomeExpenseReportInput) (*GetIncomeExpenseReportResult, error) {
	input.PeriodStart = calendarDate(input.PeriodStart, s.location)
	input.PeriodEnd = calendarDate(input.PeriodEnd, s.location)

	if err := input.Validate(); err != nil {
		return nil, err
	}

	currency, err := valueObject.NewCurrency(valueObject.CurrencyCode(input.Currency))
	if err != nil {
		return nil, apperror.Invalid(err.Error())
	}

	if err := s.checkAccounts(ctx, input.UserID, input.AccountIDs, currency); err != nil {
		return nil, err
	}

	report, err := s.reportQueryService.GetIncomeExpenseReport(ctx, query.GetIncomeExpenseReportCriteria{
		UserID:      input.UserID,
		Currency:    currency.Code,
		Location:    *s.location,
		PeriodStart: input.PeriodStart,
		PeriodEnd:   input.PeriodEnd,
		AccountIDs:  input.AccountIDs,
	})
	if err != nil {
		return nil, err
	}

	items := make([]IncomeExpenseReportItem, len(report.Items))
	for i, item := range report.Items {
		items[i] = IncomeExpenseReportItem{
			Month:   item.Month,
			Income:  item.Income,
			Expense: item.Expense,
		}
	}

	return &GetIncomeExpenseReportResult{Items: items}, nil
}

func (i *GetIncomeExpenseReportInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Currency == "", "currency", "is required")
	v.AddIf(i.PeriodStart != nil && i.PeriodEnd != nil && i.PeriodStart.After(*i.PeriodEnd), "periodStart", "must not be after periodEnd")
	v.AddIf(i.PeriodStart != nil && i.PeriodEnd != nil && periodExceedsLimit(*i.PeriodStart, *i.PeriodEnd), "periodEnd", "period must not exceed 100 years")
	v.AddIf(len(i.AccountIDs) > MaxReportAccountIDs, "accountIds", "must not contain more than 100 items")
	return v.ErrorOrNil()
}
