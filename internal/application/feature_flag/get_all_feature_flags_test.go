package feature_flag_test

import (
	"context"
	"errors"
	"testing"

	"finance-app-go-demo/internal/application/feature_flag"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetAllFeatureFlags(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(flags *repomock.MockFeatureFlagRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *feature_flag.GetAllFeatureFlagsResult)
	}{
		{
			name: "returns all flags as stored",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetAll(mock.Anything).Return([]model.FeatureFlag{
					{ID: uuid.New(), Key: "ai-receipts", Description: "AI receipt parsing", Enabled: false},
					{ID: uuid.New(), Key: "registration", Enabled: true},
				}, nil)
			},
			wantResult: func(t *testing.T, result *feature_flag.GetAllFeatureFlagsResult) {
				require.Len(t, result.FeatureFlags, 2)
				assert.Equal(t, "ai-receipts", result.FeatureFlags[0].Key)
				assert.False(t, result.FeatureFlags[0].Enabled)
				assert.Equal(t, "registration", result.FeatureFlags[1].Key)
				assert.True(t, result.FeatureFlags[1].Enabled)
			},
		},
		{
			name: "no flags",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetAll(mock.Anything).Return([]model.FeatureFlag{}, nil)
			},
			wantResult: func(t *testing.T, result *feature_flag.GetAllFeatureFlagsResult) {
				assert.Empty(t, result.FeatureFlags)
			},
		},
		{
			name: "repository error propagates",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetAll(mock.Anything).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := repomock.NewMockFeatureFlagRepository(t)
			tt.setupMocks(flags)

			service := feature_flag.NewService(flags, repomock.NewMockFeatureFlagOverrideRepository(t))

			result, err := service.GetAllFeatureFlags(context.Background())

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
