package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/featureflag"
	"finance-app-go-demo/ent/featureflagoverride"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"

	"github.com/google/uuid"
)

func toFeatureFlagModel(f *ent.FeatureFlag) *model.FeatureFlag {
	if f == nil {
		return nil
	}

	return &model.FeatureFlag{
		ID:          f.ID,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		Key:         f.Key,
		Description: f.Description,
		Enabled:     f.Enabled,
	}
}

type featureFlagRepository struct {
	client *ent.Client
}

func NewFeatureFlagRepository(client *ent.Client) repository.FeatureFlagRepository {
	return &featureFlagRepository{client: client}
}

func (r *featureFlagRepository) Create(ctx context.Context, f *model.FeatureFlag) (*model.FeatureFlag, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlag.Create().
		SetKey(f.Key).
		SetDescription(f.Description).
		SetEnabled(f.Enabled).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toFeatureFlagModel(entity), nil
}

func (r *featureFlagRepository) Update(ctx context.Context, f *model.FeatureFlag) (*model.FeatureFlag, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlag.UpdateOneID(f.ID).
		SetKey(f.Key).
		SetDescription(f.Description).
		SetEnabled(f.Enabled).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return toFeatureFlagModel(entity), nil
}

func (r *featureFlagRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.FeatureFlag, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlag.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toFeatureFlagModel(entity), nil
}

func (r *featureFlagRepository) GetByKey(ctx context.Context, key string) (*model.FeatureFlag, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlag.Query().
		Where(featureflag.Key(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toFeatureFlagModel(entity), nil
}

func (r *featureFlagRepository) GetAll(ctx context.Context) ([]model.FeatureFlag, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.FeatureFlag.Query().
		Order(ent.Asc(featureflag.FieldKey)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.FeatureFlag, len(entities))

	for i, e := range entities {
		result[i] = *toFeatureFlagModel(e)
	}

	return result, nil
}

func (r *featureFlagRepository) IsEnabled(ctx context.Context, key string) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlag.Query().
		Where(featureflag.Key(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return entity.Enabled, nil
}

func (r *featureFlagRepository) IsEnabledForUser(ctx context.Context, key string, userID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.FeatureFlag.Query().
		Where(featureflag.Key(key)).
		WithOverrides(func(q *ent.FeatureFlagOverrideQuery) {
			q.Where(featureflagoverride.UserIDEQ(userID))
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	if len(entity.Edges.Overrides) > 0 {
		return entity.Edges.Overrides[0].Enabled, nil
	}

	return entity.Enabled, nil
}

func (r *featureFlagRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.FeatureFlag.Delete().
		Where(featureflag.IDEQ(ID)).
		Exec(ctx)

	return err
}
