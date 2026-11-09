package feature_flag_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpfeatureflag "finance-app-go-demo/internal/adapter/http/feature_flag"
	"finance-app-go-demo/internal/application/feature_flag"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GetAllFeatureFlags(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(flags *repomock.MockFeatureFlagRepository)
		wantStatus int
		wantBody   string
	}{
		{
			name: "returns global values without descriptions",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetAll(mock.Anything).Return([]model.FeatureFlag{
					{ID: uuid.New(), Key: "registration", Description: "Open registration", Enabled: true},
					{ID: uuid.New(), Key: "ai-receipts", Enabled: false},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"featureFlags":[{"key":"registration","enabled":true},{"key":"ai-receipts","enabled":false}]}`,
		},
		{
			name: "empty list is an array, not null",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetAll(mock.Anything).Return(nil, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"featureFlags":[]}`,
		},
		{
			name: "repository error returns 500",
			setupMocks: func(flags *repomock.MockFeatureFlagRepository) {
				flags.EXPECT().GetAll(mock.Anything).Return(nil, errors.New("connection refused"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := repomock.NewMockFeatureFlagRepository(t)
			tt.setupMocks(flags)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			h := httpfeatureflag.NewHandler(feature_flag.NewService(flags, repomock.NewMockFeatureFlagOverrideRepository(t)))
			r.GET("/feature-flag/global", h.GetAllFeatureFlags)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/feature-flag/global", nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, rec.Body.String())
			}
		})
	}
}
