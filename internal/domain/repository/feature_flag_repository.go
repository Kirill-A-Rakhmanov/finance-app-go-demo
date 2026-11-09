package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type FeatureFlagRepository interface {
	Create(ctx context.Context, flag *model.FeatureFlag) (*model.FeatureFlag, error)
	Update(ctx context.Context, flag *model.FeatureFlag) (*model.FeatureFlag, error)
	GetByID(ctx context.Context, ID uuid.UUID) (*model.FeatureFlag, error)
	GetByKey(ctx context.Context, key string) (*model.FeatureFlag, error)
	GetAll(ctx context.Context) ([]model.FeatureFlag, error)
	IsEnabled(ctx context.Context, key string) (bool, error)
	IsEnabledForUser(ctx context.Context, key string, userID uuid.UUID) (bool, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
