package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/featureflagoverride"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"

	"github.com/google/uuid"
)

func toFeatureFlagOverrideModel(o *ent.FeatureFlagOverride) *model.FeatureFlagOverride {
	if o == nil {
		return nil
	}

	return &model.FeatureFlagOverride{
		ID:            o.ID,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
		FeatureFlagID: o.FeatureFlagID,
		UserID:        o.UserID,
		Enabled:       o.Enabled,
	}
}

type featureFlagOverrideRepository struct {
	client *ent.Client
}

func NewFeatureFlagOverrideRepository(client *ent.Client) repository.FeatureFlagOverrideRepository {
	return &featureFlagOverrideRepository{client: client}
}

func (r *featureFlagOverrideRepository) Create(ctx context.Context, o *model.FeatureFlagOverride) (*model.FeatureFlagOverride, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlagOverride.Create().
		SetFeatureFlagID(o.FeatureFlagID).
		SetUserID(o.UserID).
		SetEnabled(o.Enabled).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toFeatureFlagOverrideModel(entity), nil
}

func (r *featureFlagOverrideRepository) Update(ctx context.Context, o *model.FeatureFlagOverride) (*model.FeatureFlagOverride, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlagOverride.UpdateOneID(o.ID).
		SetFeatureFlagID(o.FeatureFlagID).
		SetUserID(o.UserID).
		SetEnabled(o.Enabled).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toFeatureFlagOverrideModel(entity), nil
}

func (r *featureFlagOverrideRepository) GetByFeatureFlagIDAndUserID(ctx context.Context, featureFlagID uuid.UUID, userID uuid.UUID) (*model.FeatureFlagOverride, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlagOverride.Query().
		Where(
			featureflagoverride.FeatureFlagIDEQ(featureFlagID),
			featureflagoverride.UserIDEQ(userID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toFeatureFlagOverrideModel(entity), nil
}

func (r *featureFlagOverrideRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.FeatureFlagOverride, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.FeatureFlagOverride.Query().
		Where(featureflagoverride.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.FeatureFlagOverride, len(entities))

	for i, e := range entities {
		result[i] = *toFeatureFlagOverrideModel(e)
	}

	return result, nil
}

func (r *featureFlagOverrideRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.FeatureFlagOverride.Delete().
		Where(featureflagoverride.IDEQ(ID)).
		Exec(ctx)

	return err
}
