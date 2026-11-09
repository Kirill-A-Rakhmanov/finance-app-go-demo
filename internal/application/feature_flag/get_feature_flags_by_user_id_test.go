package feature_flag_test

import (
	"context"
	"errors"
	"testing"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/feature_flag"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetFeatureFlagsByUserID(t *testing.T) {
	userID := uuid.New()
	enabledID := uuid.New()
	disabledID := uuid.New()
	otherID := uuid.New()

	allFlags := []model.FeatureFlag{
		{ID: enabledID, Key: "enabled-flag", Enabled: true},
		{ID: disabledID, Key: "disabled-flag", Enabled: false},
		{ID: otherID, Key: "other-flag", Enabled: false},
	}

	effective := func(result *feature_flag.GetFeatureFlagsByUserIDResult) map[string]bool {
		values := make(map[string]bool, len(result.FeatureFlags))
		for _, flag := range result.FeatureFlags {
			values[flag.Key] = flag.Enabled
		}

		return values
	}

	type mocks struct {
		flags     *repomock.MockFeatureFlagRepository
		overrides *repomock.MockFeatureFlagOverrideRepository
	}

	tests := []struct {
		name       string
		input      feature_flag.GetFeatureFlagsByUserIDInput
		setupMocks func(m mocks)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *feature_flag.GetFeatureFlagsByUserIDResult)
	}{
		{
			name:  "without overrides the global values apply",
			input: feature_flag.GetFeatureFlagsByUserIDInput{UserID: userID},
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(allFlags, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.FeatureFlagOverride{}, nil)
			},
			wantResult: func(t *testing.T, result *feature_flag.GetFeatureFlagsByUserIDResult) {
				assert.Equal(t, map[string]bool{"enabled-flag": true, "disabled-flag": false, "other-flag": false}, effective(result))
			},
		},
		{
			name:  "an override wins in both directions",
			input: feature_flag.GetFeatureFlagsByUserIDInput{UserID: userID},
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(allFlags, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.FeatureFlagOverride{
					{FeatureFlagID: enabledID, UserID: userID, Enabled: false},
					{FeatureFlagID: disabledID, UserID: userID, Enabled: true},
				}, nil)
			},
			wantResult: func(t *testing.T, result *feature_flag.GetFeatureFlagsByUserIDResult) {
				assert.Equal(t, map[string]bool{"enabled-flag": false, "disabled-flag": true, "other-flag": false}, effective(result))
			},
		},
		{
			name:  "keeps the order of the repository",
			input: feature_flag.GetFeatureFlagsByUserIDInput{UserID: userID},
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(allFlags, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, userID).Return(nil, nil)
			},
			wantResult: func(t *testing.T, result *feature_flag.GetFeatureFlagsByUserIDResult) {
				require.Len(t, result.FeatureFlags, 3)
				assert.Equal(t, "enabled-flag", result.FeatureFlags[0].Key)
				assert.Equal(t, "disabled-flag", result.FeatureFlags[1].Key)
				assert.Equal(t, "other-flag", result.FeatureFlags[2].Key)
			},
		},
		{
			name:  "no flags",
			input: feature_flag.GetFeatureFlagsByUserIDInput{UserID: userID},
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return([]model.FeatureFlag{}, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.FeatureFlagOverride{}, nil)
			},
			wantResult: func(t *testing.T, result *feature_flag.GetFeatureFlagsByUserIDResult) {
				assert.NotNil(t, result.FeatureFlags)
				assert.Empty(t, result.FeatureFlags)
			},
		},
		{
			name:       "validation error - empty user id",
			input:      feature_flag.GetFeatureFlagsByUserIDInput{UserID: uuid.Nil},
			setupMocks: func(m mocks) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "flag repository error propagates",
			input: feature_flag.GetFeatureFlagsByUserIDInput{UserID: userID},
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
		{
			name:  "override repository error propagates",
			input: feature_flag.GetFeatureFlagsByUserIDInput{UserID: userID},
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(allFlags, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks{
				flags:     repomock.NewMockFeatureFlagRepository(t),
				overrides: repomock.NewMockFeatureFlagOverrideRepository(t),
			}
			tt.setupMocks(m)

			service := feature_flag.NewService(m.flags, m.overrides)

			result, err := service.GetFeatureFlagsByUserID(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}
