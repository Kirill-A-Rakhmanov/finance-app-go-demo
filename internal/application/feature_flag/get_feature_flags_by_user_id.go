package feature_flag

import (
	"context"

	apperror "finance-app-go-demo/internal/application/error"

	"github.com/google/uuid"
)

type GetFeatureFlagsByUserIDInput struct {
	UserID uuid.UUID
}

type UserFeatureFlag struct {
	Key     string
	Enabled bool
}

type GetFeatureFlagsByUserIDResult struct {
	FeatureFlags []UserFeatureFlag
}

func (s *Service) GetFeatureFlagsByUserID(
	ctx context.Context,
	input GetFeatureFlagsByUserIDInput,
) (*GetFeatureFlagsByUserIDResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	flags, err := s.featureFlagRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	overrides, err := s.featureFlagOverrideRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	overridden := make(map[uuid.UUID]bool, len(overrides))
	for _, override := range overrides {
		overridden[override.FeatureFlagID] = override.Enabled
	}

	result := make([]UserFeatureFlag, len(flags))
	for i, flag := range flags {
		enabled := flag.Enabled
		if value, ok := overridden[flag.ID]; ok {
			enabled = value
		}

		result[i] = UserFeatureFlag{Key: flag.Key, Enabled: enabled}
	}

	return &GetFeatureFlagsByUserIDResult{
		FeatureFlags: result,
	}, nil
}

func (i *GetFeatureFlagsByUserIDInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	return v.ErrorOrNil()
}
