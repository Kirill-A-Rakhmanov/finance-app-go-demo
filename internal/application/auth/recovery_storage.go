package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Recovery struct {
	Code      string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type RecoveryStorage interface {
	CreateRecovery(ctx context.Context, userID uuid.UUID) (*Recovery, error)
	GetRecovery(ctx context.Context, code string) (*Recovery, error)
	DeleteRecovery(ctx context.Context, code string) error
	AcquireCooldown(ctx context.Context, email string, cooldown time.Duration) (bool, error)
}
