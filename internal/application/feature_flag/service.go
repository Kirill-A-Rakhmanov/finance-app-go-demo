package feature_flag

import "finance-app-go-demo/internal/domain/repository"

type Service struct {
	featureFlagRepo         repository.FeatureFlagRepository
	featureFlagOverrideRepo repository.FeatureFlagOverrideRepository
}

func NewService(
	featureFlagRepo repository.FeatureFlagRepository,
	featureFlagOverrideRepo repository.FeatureFlagOverrideRepository,
) *Service {
	return &Service{
		featureFlagRepo:         featureFlagRepo,
		featureFlagOverrideRepo: featureFlagOverrideRepo,
	}
}
