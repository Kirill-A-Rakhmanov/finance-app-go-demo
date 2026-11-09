package query

import (
	"context"
	"finance-app-go-demo/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type TransactionSortBy string

const (
	TransactionSortByDate   TransactionSortBy = "date"
	TransactionSortByAmount TransactionSortBy = "amount"
)

type GetTransactionCriteria struct {
	UserID                 uuid.UUID
	AccountID              *uuid.UUID
	TransactionCategoryIDs []uuid.UUID
	PeriodStart            *time.Time
	PeriodEnd              *time.Time
	Limit                  int
	Offset                 int
	SortBy                 *TransactionSortBy
	SortDirection          *SortDirection
}

type GetTransactionsResult struct {
	Total        int
	Order        []uuid.UUID
	Transactions []model.Transaction
}

type TransactionQueryService interface {
	GetTransactions(ctx context.Context, criteria GetTransactionCriteria) (*GetTransactionsResult, error)
}
