package request

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

var sessionKey = contextKey{}

type Session struct {
	UserID    uuid.UUID `json:"userId"`
	SessionID string    `json:"sessionId"`
}

func WithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func CurrentSession(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(sessionKey).(Session)

	return session, ok
}
