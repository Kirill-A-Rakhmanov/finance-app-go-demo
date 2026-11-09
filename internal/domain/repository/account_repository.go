package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, b *model.Account) (*model.Account, error)
	Update(ctx context.Context, b *model.Account) (*model.Account, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.Account, error)
	ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Account, error)
	GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.Account, error)
	GetByIDsAndUserID(ctx context.Context, IDs []uuid.UUID, userID uuid.UUID) ([]model.Account, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
