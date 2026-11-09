package feature_flag

import (
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/adapter/http/response"
	"finance-app-go-demo/internal/application/feature_flag"
	"fmt"

	"github.com/gin-gonic/gin"
)

type FeatureFlag struct {
	Key     string `json:"key" validate:"required" example:"ai-receipts"`
	Enabled bool   `json:"enabled" validate:"required"`
}

type GetFeatureFlagsResponse struct {
	FeatureFlags []FeatureFlag `json:"featureFlags" validate:"required"`
}

// @Summary Get feature flags
// @Description Feature flags with the value effective for the current user: the user's own override when there is one, the global value otherwise
// @Tags feature-flag
// @Accept json
// @Produce json
// @Success 200 {object} GetFeatureFlagsResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /feature-flag/list [get]
func (h *Handler) GetFeatureFlags(c *gin.Context) {
	ctx := c.Request.Context()

	session, ok := request.CurrentSession(ctx)
	if !ok {
		response.Unauthorized(c.Writer, fmt.Errorf("user not authorized"))
		return
	}

	result, err := h.featureFlagService.GetFeatureFlagsByUserID(ctx, feature_flag.GetFeatureFlagsByUserIDInput{
		UserID: session.UserID,
	})
	if err != nil {
		response.FromError(c.Writer, err)
		return
	}

	flags := make([]FeatureFlag, len(result.FeatureFlags))
	for i, flag := range result.FeatureFlags {
		flags[i] = FeatureFlag{Key: flag.Key, Enabled: flag.Enabled}
	}

	response.OK(c.Writer, GetFeatureFlagsResponse{FeatureFlags: flags})
}
