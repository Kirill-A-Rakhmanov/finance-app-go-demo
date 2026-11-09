package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type SessionStorage interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error)
	GetSession(ctx context.Context, ID string) (*Session, error)
	DeleteSession(ctx context.Context, ID string) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, exceptSessionID string) error
	ShortenSession(ctx context.Context, ID string, ttl time.Duration) error
}
