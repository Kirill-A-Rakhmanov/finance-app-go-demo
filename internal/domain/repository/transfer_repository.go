package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type TransferRepository interface {
	Create(ctx context.Context, t *model.Transfer) (*model.Transfer, error)
	Update(ctx context.Context, t *model.Transfer) (*model.Transfer, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.Transfer, error)
	ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error)
	GetByFromAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transfer, error)
	GetByToAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transfer, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
