package model

import (
	"time"

	"github.com/google/uuid"
)

type FeatureFlagOverride struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	FeatureFlagID uuid.UUID
	UserID        uuid.UUID
	Enabled       bool
}

func NewFeatureFlagOverride(
	featureFlagID uuid.UUID,
	userID uuid.UUID,
	enabled bool,
) *FeatureFlagOverride {
	return &FeatureFlagOverride{
		FeatureFlagID: featureFlagID,
		UserID:        userID,
		Enabled:       enabled,
	}
}
