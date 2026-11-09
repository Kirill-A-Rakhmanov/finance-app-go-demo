package feature_flag

import "finance-app-go-demo/internal/application/feature_flag"

type Handler struct {
	featureFlagService *feature_flag.Service
}

func NewHandler(featureFlagService *feature_flag.Service) *Handler {
	return &Handler{
		featureFlagService: featureFlagService,
	}
}
