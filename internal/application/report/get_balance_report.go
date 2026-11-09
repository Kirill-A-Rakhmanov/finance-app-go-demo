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

type GetBalanceReportInput struct {
	UserID      uuid.UUID
	Currency    string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	AccountIDs  []uuid.UUID
}

type BalanceReportItem struct {
	DateKey string
	Balance decimal.Decimal
}

type GetBalanceReportResult struct {
	Granularity query.Granularity
	Items       []BalanceReportItem
}

func (s *Service) GetBalanceReport(ctx context.Context, input GetBalanceReportInput) (*GetBalanceReportResult, error) {
	input.PeriodStart = calendarDate(input.PeriodStart, s.location)
	input.PeriodEnd = calendarDate(input.PeriodEnd, s.location)

	if input.PeriodEnd == nil {
		now := time.Now()
		input.PeriodEnd = calendarDate(&now, s.location)
	}

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

	if input.PeriodStart == nil {
		periodStart, err := s.balanceReportStart(ctx, input, currency)
		if err != nil {
			return nil, err
		}

		input.PeriodStart = &periodStart
	}

	granularity := granularityForDays(calendarDays(*input.PeriodStart, *input.PeriodEnd))

	report, err := s.reportQueryService.GetBalanceReport(ctx, query.GetBalanceReportCriteria{
		UserID:      input.UserID,
		Currency:    currency.Code,
		Location:    *s.location,
		Granularity: granularity,
		PeriodStart: *input.PeriodStart,
		PeriodEnd:   *input.PeriodEnd,
		AccountIDs:  input.AccountIDs,
	})
	if err != nil {
		return nil, err
	}

	items := make([]BalanceReportItem, len(report.Items))
	for i, item := range report.Items {
		items[i] = BalanceReportItem{
			DateKey: item.DateKey,
			Balance: item.Balance,
		}
	}

	return &GetBalanceReportResult{
		Granularity: granularity,
		Items:       items,
	}, nil
}

func (s *Service) balanceReportStart(ctx context.Context, input GetBalanceReportInput, currency valueObject.Currency) (time.Time, error) {
	start, err := s.reportQueryService.GetBalanceReportStart(ctx, query.GetBalanceReportStartCriteria{
		UserID:     input.UserID,
		Currency:   currency.Code,
		Location:   *s.location,
		AccountIDs: input.AccountIDs,
	})
	if err != nil {
		return time.Time{}, err
	}

	latest := input.PeriodEnd.AddDate(0, 0, -(MinBalanceReportDays - 1))
	earliest := input.PeriodEnd.AddDate(-MaxReportPeriodYears, 0, 0)

	switch {
	case start.Date == nil || start.Date.After(latest):
		return latest, nil
	case start.Date.Before(earliest):
		return earliest, nil
	default:
		return *start.Date, nil
	}
}

func (i *GetBalanceReportInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Currency == "", "currency", "is required")
	v.AddIf(len(i.AccountIDs) > MaxReportAccountIDs, "accountIds", "must not contain more than 100 items")

	if i.PeriodStart != nil && i.PeriodEnd != nil {
		v.AddIf(i.PeriodStart.After(*i.PeriodEnd), "periodStart", "must not be after periodEnd")
		v.AddIf(!i.PeriodStart.After(*i.PeriodEnd) && calendarDays(*i.PeriodStart, *i.PeriodEnd) < MinBalanceReportDays, "periodEnd", "period must be at least 2 days")
		v.AddIf(periodExceedsLimit(*i.PeriodStart, *i.PeriodEnd), "periodEnd", "period must not exceed 100 years")
	}

	return v.ErrorOrNil()
}
