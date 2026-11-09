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

const (
	MaxReportAccountIDs  = 100
	MaxReportCategoryIDs = 100
)

var oneHundred = decimal.NewFromInt(100)

type GetCategoryReportInput struct {
	UserID                 uuid.UUID
	TransactionType        enum.TransactionType
	Currency               string
	PeriodStart            *time.Time
	PeriodEnd              *time.Time
	TransactionCategoryIDs []uuid.UUID
	GroupSubcategories     bool
	AccountIDs             []uuid.UUID
}

type CategoryReportItem struct {
	TransactionCategory *model.TransactionCategory
	Amount              decimal.Decimal
	Share               decimal.Decimal
}

type GetCategoryReportResult struct {
	Total decimal.Decimal
	Items []CategoryReportItem
}

func (s *Service) GetCategoryReport(ctx context.Context, input GetCategoryReportInput) (*GetCategoryReportResult, error) {
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

	if err := s.checkCategories(ctx, input.UserID, input.TransactionCategoryIDs, input.GroupSubcategories); err != nil {
		return nil, err
	}

	report, err := s.reportQueryService.GetCategoryReport(ctx, query.GetCategoryReportCriteria{
		UserID:                 input.UserID,
		TransactionType:        input.TransactionType,
		Currency:               currency.Code,
		PeriodStart:            input.PeriodStart,
		PeriodEnd:              input.PeriodEnd,
		TransactionCategoryIDs: input.TransactionCategoryIDs,
		GroupSubcategories:     input.GroupSubcategories,
		AccountIDs:             input.AccountIDs,
	})
	if err != nil {
		return nil, err
	}

	total := decimal.Zero
	for _, item := range report.Items {
		total = total.Add(item.Amount)
	}

	items := make([]CategoryReportItem, len(report.Items))
	for i, item := range report.Items {
		items[i] = CategoryReportItem{
			TransactionCategory: item.TransactionCategory,
			Amount:              item.Amount,
			Share:               calcShare(item.Amount, total),
		}
	}

	return &GetCategoryReportResult{
		Total: total,
		Items: items,
	}, nil
}

func (s *Service) checkAccounts(
	ctx context.Context,
	userID uuid.UUID,
	accountIDs []uuid.UUID,
	currency valueObject.Currency,
) error {
	if len(accountIDs) == 0 {
		return nil
	}

	requested := make(map[uuid.UUID]struct{}, len(accountIDs))
	for _, id := range accountIDs {
		requested[id] = struct{}{}
	}

	accounts, err := s.accountRepo.GetByIDsAndUserID(ctx, accountIDs, userID)
	if err != nil {
		return err
	}

	if len(accounts) != len(requested) {
		return apperror.ErrAccountNotFound
	}

	for _, account := range accounts {
		if account.Currency.Code != currency.Code {
			return ErrAccountCurrencyMismatch
		}
	}

	return nil
}

func (s *Service) checkCategories(ctx context.Context, userID uuid.UUID, transactionCategoryIDs []uuid.UUID, groupSubcategories bool) error {
	if len(transactionCategoryIDs) == 0 {
		return nil
	}

	requested := make(map[uuid.UUID]struct{}, len(transactionCategoryIDs))
	for _, id := range transactionCategoryIDs {
		requested[id] = struct{}{}
	}

	categories, err := s.transactionCategoryRepo.GetByIDsAndUserID(ctx, transactionCategoryIDs, userID)
	if err != nil {
		return err
	}

	if len(categories) != len(requested) {
		return apperror.ErrTransactionCategoryNotFound
	}

	if groupSubcategories {
		for _, category := range categories {
			if category.ParentID != nil {
				return ErrSubcategoryFilterNotAllowed
			}
		}
	}

	return nil
}

func calcShare(amount, total decimal.Decimal) decimal.Decimal {
	share := decimal.Zero
	if !total.IsZero() {
		share = amount.Mul(oneHundred).Div(total).Round(2)
	}

	return share
}

func (i *GetCategoryReportInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.TransactionType != enum.TransactionIncome && i.TransactionType != enum.TransactionExpense, "transactionType", "must be income or expense")
	v.AddIf(i.Currency == "", "currency", "is required")
	v.AddIf(i.PeriodStart != nil && i.PeriodEnd != nil && i.PeriodStart.After(*i.PeriodEnd), "periodStart", "must not be after periodEnd")
	v.AddIf(len(i.AccountIDs) > MaxReportAccountIDs, "accountIds", "must not contain more than 100 items")
	v.AddIf(len(i.TransactionCategoryIDs) > MaxReportCategoryIDs, "transactionCategoryIds", "must not contain more than 100 items")
	return v.ErrorOrNil()
}
