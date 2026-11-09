//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	infraredis "finance-app-go-demo/internal/infrastructure/redis"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func uniqueKey() string {
	return "login:" + uuid.NewString() + "@example.com"
}

func TestFailureLimiter_NotLockedBeforeTheLimit(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	key := uniqueKey()

	retryAfter, err := limiter.RetryAfter(ctx, key, 3)
	require.NoError(t, err)
	require.Zero(t, retryAfter, "an unknown key is not locked")

	for i := 0; i < 2; i++ {
		require.NoError(t, limiter.RecordFailure(ctx, key, time.Minute))
	}

	retryAfter, err = limiter.RetryAfter(ctx, key, 3)
	require.NoError(t, err)
	require.Zero(t, retryAfter)
}

func TestFailureLimiter_LocksAtTheLimitAndReportsTheRemainingTime(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	key := uniqueKey()

	for i := 0; i < 3; i++ {
		require.NoError(t, limiter.RecordFailure(ctx, key, time.Minute))
	}

	retryAfter, err := limiter.RetryAfter(ctx, key, 3)
	require.NoError(t, err)
	require.Greater(t, retryAfter, 50*time.Second)
	require.LessOrEqual(t, retryAfter, time.Minute)
}

func TestFailureLimiter_WindowStartsAtTheFirstFailureAndIsNotExtended(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	key := uniqueKey()

	require.NoError(t, limiter.RecordFailure(ctx, key, 3*time.Second))
	time.Sleep(1500 * time.Millisecond)
	require.NoError(t, limiter.RecordFailure(ctx, key, 3*time.Second))
	require.NoError(t, limiter.RecordFailure(ctx, key, 3*time.Second))

	retryAfter, err := limiter.RetryAfter(ctx, key, 3)
	require.NoError(t, err)
	require.Greater(t, retryAfter, time.Duration(0))
	require.LessOrEqual(t, retryAfter, 2*time.Second, "later failures must not push the end of the window")
}

func TestFailureLimiter_UnlocksWhenTheWindowEnds(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	key := uniqueKey()

	for i := 0; i < 2; i++ {
		require.NoError(t, limiter.RecordFailure(ctx, key, time.Second))
	}

	locked, err := limiter.RetryAfter(ctx, key, 2)
	require.NoError(t, err)
	require.Greater(t, locked, time.Duration(0))

	require.Eventually(t, func() bool {
		retryAfter, err := limiter.RetryAfter(ctx, key, 2)
		return err == nil && retryAfter == 0
	}, 5*time.Second, 100*time.Millisecond)
}

func TestFailureLimiter_ResetClearsTheCounter(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	key := uniqueKey()

	for i := 0; i < 3; i++ {
		require.NoError(t, limiter.RecordFailure(ctx, key, time.Minute))
	}
	require.NoError(t, limiter.Reset(ctx, key))

	retryAfter, err := limiter.RetryAfter(ctx, key, 3)
	require.NoError(t, err)
	require.Zero(t, retryAfter)

	require.NoError(t, limiter.Reset(ctx, uniqueKey()), "resetting an unknown key is fine")
}

func TestFailureLimiter_KeysAreIndependent(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	locked, other := uniqueKey(), uniqueKey()

	for i := 0; i < 3; i++ {
		require.NoError(t, limiter.RecordFailure(ctx, locked, time.Minute))
	}

	retryAfter, err := limiter.RetryAfter(ctx, other, 3)
	require.NoError(t, err)
	require.Zero(t, retryAfter)
}

func TestFailureLimiter_KeyIsHashedInRedis(t *testing.T) {
	ctx := context.Background()
	limiter := infraredis.NewFailureLimiter(testClient)
	key := uniqueKey()

	require.NoError(t, limiter.RecordFailure(ctx, key, time.Minute))

	keys, err := testClient.Keys(ctx, "failures:*"+key+"*").Result()
	require.NoError(t, err)
	require.Empty(t, keys, "the email must not appear in Redis keys")
}
