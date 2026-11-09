package report

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetCategoryDynamicReportInput struct {
	UserID                 uuid.UUID
	Currency               string
	PeriodStart            *time.Time
	PeriodEnd              *time.Time
	TransactionCategoryIDs []uuid.UUID
	AccountIDs             []uuid.UUID
}

type CategoryDynamicReportItem struct {
	Month                 string
	TransactionCategoryID uuid.UUID
	Amount                decimal.Decimal
}

type GetCategoryDynamicReportResult struct {
	Months                []string
	Items                 []CategoryDynamicReportItem
	TransactionCategories []model.TransactionCategory
}

func (s *Service) GetCategoryDynamicReport(ctx context.Context, input GetCategoryDynamicReportInput) (*GetCategoryDynamicReportResult, error) {
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

	if err := s.checkCategories(ctx, input.UserID, input.TransactionCategoryIDs, false); err != nil {
		return nil, err
	}

	report, err := s.reportQueryService.GetCategoryDynamicReport(ctx, query.GetCategoryDynamicReportCriteria{
		UserID:                 input.UserID,
		Currency:               currency.Code,
		PeriodStart:            input.PeriodStart,
		PeriodEnd:              input.PeriodEnd,
		Location:               *s.location,
		TransactionCategoryIDs: input.TransactionCategoryIDs,
		AccountIDs:             input.AccountIDs,
	})
	if err != nil {
		return nil, err
	}

	result := &GetCategoryDynamicReportResult{
		Months:                make([]string, 0, len(report.Items)),
		Items:                 []CategoryDynamicReportItem{},
		TransactionCategories: []model.TransactionCategory{},
	}

	seen := make(map[uuid.UUID]struct{})

	for _, month := range report.Items {
		result.Months = append(result.Months, month.Month)

		for _, category := range month.Categories {
			result.Items = append(result.Items, CategoryDynamicReportItem{
				Month:                 month.Month,
				TransactionCategoryID: category.TransactionCategory.ID,
				Amount:                category.Amount,
			})

			if _, ok := seen[category.TransactionCategory.ID]; !ok {
				seen[category.TransactionCategory.ID] = struct{}{}
				result.TransactionCategories = append(result.TransactionCategories, category.TransactionCategory)
			}
		}
	}

	return result, nil
}

func (i *GetCategoryDynamicReportInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Currency == "", "currency", "is required")
	v.AddIf(i.PeriodStart != nil && i.PeriodEnd != nil && i.PeriodStart.After(*i.PeriodEnd), "periodStart", "must not be after periodEnd")
	v.AddIf(i.PeriodStart != nil && i.PeriodEnd != nil && periodExceedsLimit(*i.PeriodStart, *i.PeriodEnd), "periodEnd", "period must not exceed 100 years")
	v.AddIf(len(i.AccountIDs) > MaxReportAccountIDs, "accountIds", "must not contain more than 100 items")
	v.AddIf(len(i.TransactionCategoryIDs) == 0, "transactionCategoryIds", "must contain at least 1 item")
	v.AddIf(len(i.TransactionCategoryIDs) > MaxReportCategoryIDs, "transactionCategoryIds", "must not contain more than 100 items")
	return v.ErrorOrNil()
}
