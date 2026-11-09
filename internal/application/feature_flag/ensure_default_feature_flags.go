package feature_flag

import (
	"context"

	"finance-app-go-demo/internal/domain/model"
)

func (s *Service) EnsureDefaultFeatureFlags(ctx context.Context) ([]string, error) {
	created := []string{}

	for _, def := range model.DefaultFeatureFlags {
		existing, err := s.featureFlagRepo.GetByKey(ctx, def.Key)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			continue
		}

		if _, err := s.featureFlagRepo.Create(ctx, model.NewFeatureFlag(def.Key, def.Description, def.Enabled)); err != nil {
			return nil, err
		}

		created = append(created, def.Key)
	}

	return created, nil
}
