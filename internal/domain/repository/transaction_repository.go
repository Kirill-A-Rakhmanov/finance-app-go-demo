package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionRepository interface {
	Create(ctx context.Context, t *model.Transaction) (*model.Transaction, error)
	BulkCreate(ctx context.Context, ts []model.Transaction) ([]model.Transaction, error)
	Update(ctx context.Context, t *model.Transaction) (*model.Transaction, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.Transaction, error)
	ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transaction, error)
	GetByIDAndAccountID(ctx context.Context, ID uuid.UUID, accountID uuid.UUID) (*model.Transaction, error)
	GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.Transaction, error)
	GetByTransactionCategoryID(ctx context.Context, transactionCategoryID uuid.UUID) ([]model.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Transaction, error)
	GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error)
	GetBalanceByUserID(ctx context.Context, userID uuid.UUID) ([]valueObject.Balance, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
