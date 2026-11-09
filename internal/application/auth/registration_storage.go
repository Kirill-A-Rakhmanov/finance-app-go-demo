package auth

import (
	"context"
	"time"
)

type PendingRegistration struct {
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
}

type RegistrationStorage interface {
	AcquireCooldown(ctx context.Context, email string, cooldown time.Duration) (bool, error)
	Create(ctx context.Context, registration PendingRegistration) (token string, err error)
	Get(ctx context.Context, token string) (*PendingRegistration, error)
	Delete(ctx context.Context, token string) error
}
