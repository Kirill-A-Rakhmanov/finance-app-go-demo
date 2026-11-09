package model

import (
	"time"

	"github.com/google/uuid"
)

const FeatureFlagRegistration = "registration"

var DefaultFeatureFlags = []FeatureFlag{
	{
		Key:         FeatureFlagRegistration,
		Description: "Open registration of new users",
		Enabled:     true,
	},
}

type FeatureFlag struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time

	Key         string
	Description string
	Enabled     bool
}

func NewFeatureFlag(
	key string,
	description string,
	enabled bool,
) *FeatureFlag {
	return &FeatureFlag{
		Key:         key,
		Description: description,
		Enabled:     enabled,
	}
}
