package user

import (
	"context"

	"github.com/google/uuid"
)

type SessionRevoker interface {
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID, exceptSessionID string) error
}
