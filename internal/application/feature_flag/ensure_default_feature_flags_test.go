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

func TestService_EnsureDefaultFeatureFlags(t *testing.T) {
	registration := model.DefaultFeatureFlags[0]

	tests := []struct {
		name       string
		setupMocks func(flags *repomock.MockFeatureFlagRepository)
		wantErr    func(t *testing.T, err error)
		wantKeys   []string
	}{
		{
			name: "creates missing flag with default values",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetByKey(mock.Anything, registration.Key).Return(nil, nil)
				flags.EXPECT().Create(mock.Anything, mock.MatchedBy(func(f *model.FeatureFlag) bool {
					return f.Key == registration.Key &&
						f.Description == registration.Description &&
						f.Enabled == registration.Enabled
				})).Return(&model.FeatureFlag{ID: uuid.New(), Key: registration.Key}, nil)
			},
			wantKeys: []string{registration.Key},
		},
		{
			name: "existing flag is left untouched",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetByKey(mock.Anything, registration.Key).
					Return(&model.FeatureFlag{ID: uuid.New(), Key: registration.Key, Enabled: false}, nil)
			},
			wantKeys: []string{},
		},
		{
			name: "lookup error propagates",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetByKey(mock.Anything, registration.Key).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
		{
			name: "create error propagates",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetByKey(mock.Anything, registration.Key).Return(nil, nil)
				flags.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := repomock.NewMockFeatureFlagRepository(t)
			tt.setupMocks(flags)

			service := feature_flag.NewService(flags, repomock.NewMockFeatureFlagOverrideRepository(t))

			keys, err := service.EnsureDefaultFeatureFlags(context.Background())

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, keys)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantKeys, keys)
		})
	}
}
