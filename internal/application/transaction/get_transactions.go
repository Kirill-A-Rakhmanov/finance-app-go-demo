package transaction

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultTransactionsLimit  = 50
	MaxTransactionsLimit      = 200
	MaxTransactionCategoryIDs = 100
)

type GetTransactionsInput struct {
	UserID                 uuid.UUID
	AccountID              *uuid.UUID
	TransactionCategoryIDs []uuid.UUID
	PeriodStart            *time.Time
	PeriodEnd              *time.Time
	Limit                  int
	Offset                 int
	SortBy                 *query.TransactionSortBy
	SortDirection          *query.SortDirection
}

type GetTransactionsResult struct {
	Total        int
	Limit        int
	Offset       int
	Order        []uuid.UUID
	Transactions []model.Transaction
}

func (s *Service) GetTransactions(ctx context.Context, input GetTransactionsInput) (*GetTransactionsResult, error) {
	input.applyDefaults()

	if err := input.Validate(); err != nil {
		return nil, err
	}

	result, err := s.transactionQueryService.GetTransactions(ctx, query.GetTransactionCriteria{
		UserID:                 input.UserID,
		AccountID:              input.AccountID,
		TransactionCategoryIDs: input.TransactionCategoryIDs,
		PeriodStart:            input.PeriodStart,
		PeriodEnd:              input.PeriodEnd,
		Limit:                  input.Limit,
		Offset:                 input.Offset,
		SortBy:                 input.SortBy,
		SortDirection:          input.SortDirection,
	})
	if err != nil {
		return nil, err
	}

	return &GetTransactionsResult{
		Total:        result.Total,
		Limit:        input.Limit,
		Offset:       input.Offset,
		Order:        result.Order,
		Transactions: result.Transactions,
	}, nil
}

func (i *GetTransactionsInput) applyDefaults() {
	if i.Limit == 0 {
		i.Limit = DefaultTransactionsLimit
	}
}

func (i *GetTransactionsInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.Limit < 0, "limit", "must not be negative")
	v.AddIf(i.Limit > MaxTransactionsLimit, "limit", "must not exceed 200")
	v.AddIf(i.Offset < 0, "offset", "must not be negative")
	v.AddIf(len(i.TransactionCategoryIDs) > MaxTransactionCategoryIDs, "transactionCategoryIds", "must not contain more than 100 items")
	return v.ErrorOrNil()
}
