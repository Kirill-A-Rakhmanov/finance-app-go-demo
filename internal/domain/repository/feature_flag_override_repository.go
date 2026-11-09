package repository

import (
	"context"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

type FeatureFlagOverrideRepository interface {
	Create(ctx context.Context, override *model.FeatureFlagOverride) (*model.FeatureFlagOverride, error)
	Update(ctx context.Context, override *model.FeatureFlagOverride) (*model.FeatureFlagOverride, error)
	GetByFeatureFlagIDAndUserID(ctx context.Context, featureFlagID uuid.UUID, userID uuid.UUID) (*model.FeatureFlagOverride, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.FeatureFlagOverride, error)
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}
