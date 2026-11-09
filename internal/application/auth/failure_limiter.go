package auth

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"time"

	"github.com/google/uuid"
)

type FailureLimiter interface {
	RetryAfter(ctx context.Context, key string, maxFailures int) (time.Duration, error)
	RecordFailure(ctx context.Context, key string, window time.Duration) error
	Reset(ctx context.Context, key string) error
}

func EnsureNotLocked(ctx context.Context, limiter FailureLimiter, key string) error {
	retryAfter, err := limiter.RetryAfter(ctx, key, MaxFailedAttempts)
	if err != nil {
		return err
	}
	if retryAfter > 0 {
		return apperror.TooManyRequests("too many failed attempts, try again later", retryAfter)
	}

	return nil
}

func LoginFailuresKey(email string) string {
	return "login:" + email
}

func ChangePasswordFailuresKey(userID uuid.UUID) string {
	return "change-password:" + userID.String()
}
