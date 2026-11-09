package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type CounterpartyRepository interface {
	Create(ctx context.Context, category *model.Counterparty) (*model.Counterparty, error)
	Update(ctx context.Context, category *model.Counterparty) (*model.Counterparty, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.Counterparty, error)
	GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.Counterparty, error)
	GetByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (*model.Counterparty, error)
	GetByIDsAndUserID(ctx context.Context, IDs []uuid.UUID, userID uuid.UUID) ([]model.Counterparty, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Counterparty, error)
	ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error)
	ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
