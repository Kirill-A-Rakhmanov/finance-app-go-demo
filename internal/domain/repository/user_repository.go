package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) (*model.User, error)
	Update(ctx context.Context, u *model.User) (*model.User, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.User, error)
	ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
