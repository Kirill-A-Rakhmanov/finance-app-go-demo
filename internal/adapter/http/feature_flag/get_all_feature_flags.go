package feature_flag

import (
	"finance-app-go-demo/internal/adapter/http/response"

	"github.com/gin-gonic/gin"
)

type GetAllFeatureFlagsResponse struct {
	FeatureFlags []FeatureFlag `json:"featureFlags" validate:"required"`
}

// @Summary Get global feature flags
// @Description Global values of all feature flags, without user overrides
// @Tags feature-flag
// @Accept json
// @Produce json
// @Success 200 {object} GetAllFeatureFlagsResponse
// @Router /feature-flag/global [get]
func (h *Handler) GetAllFeatureFlags(c *gin.Context) {
	result, err := h.featureFlagService.GetAllFeatureFlags(c.Request.Context())
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	flags := make([]FeatureFlag, len(result.FeatureFlags))
	for i, flag := range result.FeatureFlags {
		flags[i] = FeatureFlag{Key: flag.Key, Enabled: flag.Enabled}
	}

	response.OK(c.Writer, GetAllFeatureFlagsResponse{FeatureFlags: flags})
}
