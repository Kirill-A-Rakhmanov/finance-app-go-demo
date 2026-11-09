package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type TransactionCategoryRepository interface {
	Create(ctx context.Context, category *model.TransactionCategory) (*model.TransactionCategory, error)
	BulkCreate(ctx context.Context, categories []model.TransactionCategory) ([]model.TransactionCategory, error)
	Update(ctx context.Context, category *model.TransactionCategory) (*model.TransactionCategory, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.TransactionCategory, error)
	GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.TransactionCategory, error)
	GetByUserIDAndTypeAndName(ctx context.Context, userID uuid.UUID, type_ enum.TransactionCategoryType, name string) (*model.TransactionCategory, error)
	GetByIDsAndUserID(ctx context.Context, IDs []uuid.UUID, userID uuid.UUID) ([]model.TransactionCategory, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.TransactionCategory, error)
	GetAllByParentID(ctx context.Context, parentID uuid.UUID) ([]model.TransactionCategory, error)
	ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error)
	ExistsByUserIDAndTypeAndName(ctx context.Context, userID uuid.UUID, type_ enum.TransactionCategoryType, name string) (bool, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
