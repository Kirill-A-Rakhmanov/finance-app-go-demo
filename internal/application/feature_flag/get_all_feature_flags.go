package feature_flag

import (
	"context"

	"finance-app-go-demo/internal/domain/model"
)

type GetAllFeatureFlagsResult struct {
	FeatureFlags []model.FeatureFlag
}

func (s *Service) GetAllFeatureFlags(ctx context.Context) (*GetAllFeatureFlagsResult, error) {
	flags, err := s.featureFlagRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return &GetAllFeatureFlagsResult{
		FeatureFlags: flags,
	}, nil
}
