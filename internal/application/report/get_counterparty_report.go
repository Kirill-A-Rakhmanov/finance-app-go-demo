package report

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const MaxReportCounterpartyIDs = 100

type GetCounterpartyReportInput struct {
	UserID          uuid.UUID
	TransactionType enum.TransactionType
	Currency        string
	PeriodStart     *time.Time
	PeriodEnd       *time.Time
	CounterpartyIDs []uuid.UUID
	AccountIDs      []uuid.UUID
}

type CounterpartyReportItem struct {
	Counterparty *model.Counterparty
	Amount       decimal.Decimal
	Share        decimal.Decimal
}

type GetCounterpartyReportResult struct {
	Total decimal.Decimal
	Items []CounterpartyReportItem
}

func (s *Service) GetCounterpartyReport(ctx context.Context, input GetCounterpartyReportInput) (*GetCounterpartyReportResult, error) {
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

	if err := s.checkCounterparties(ctx, input.UserID, input.CounterpartyIDs); err != nil {
		return nil, err
	}

	report, err := s.reportQueryService.GetCounterpartyReport(ctx, query.GetCounterpartyReportCriteria{
		UserID:          input.UserID,
		TransactionType: input.TransactionType,
		Currency:        currency.Code,
		PeriodStart:     input.PeriodStart,
		PeriodEnd:       input.PeriodEnd,
		CounterpartyIDs: input.CounterpartyIDs,
		AccountIDs:      input.AccountIDs,
	})
	if err != nil {
		return nil, err
	}

	total := decimal.Zero
	for _, item := range report.Items {
		total = total.Add(item.Amount)
	}

	items := make([]CounterpartyReportItem, len(report.Items))
	for i, item := range report.Items {
		items[i] = CounterpartyReportItem{
			Counterparty: item.Counterparty,
			Amount:       item.Amount,
			Share:        calcShare(item.Amount, total),
		}
	}

	return &GetCounterpartyReportResult{
		Total: total,
		Items: items,
	}, nil
}

func (s *Service) checkCounterparties(ctx context.Context, userID uuid.UUID, counterpartyIDs []uuid.UUID) error {
	if len(counterpartyIDs) == 0 {
		return nil
	}

	requested := make(map[uuid.UUID]struct{}, len(counterpartyIDs))
	for _, id := range counterpartyIDs {
		requested[id] = struct{}{}
	}

	counterparties, err := s.counterpartyRepo.GetByIDsAndUserID(ctx, counterpartyIDs, userID)
	if err != nil {
		return err
	}

	if len(counterparties) != len(requested) {
		return apperror.ErrCounterpartyNotFound
	}

	return nil
}

func (i *GetCounterpartyReportInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.TransactionType != enum.TransactionIncome && i.TransactionType != enum.TransactionExpense, "transactionType", "must be income or expense")
	v.AddIf(i.Currency == "", "currency", "is required")
	v.AddIf(i.PeriodStart != nil && i.PeriodEnd != nil && i.PeriodStart.After(*i.PeriodEnd), "periodStart", "must not be after periodEnd")
	v.AddIf(len(i.AccountIDs) > MaxReportAccountIDs, "accountIds", "must not contain more than 100 items")
	v.AddIf(len(i.CounterpartyIDs) > MaxReportCounterpartyIDs, "counterpartyIds", "must not contain more than 100 items")
	return v.ErrorOrNil()
}
