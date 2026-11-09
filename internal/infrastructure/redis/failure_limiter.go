package redis

import (
	"context"
	"finance-app-go-demo/internal/application/auth"
	"time"

	"github.com/redis/go-redis/v9"
)

type failureLimiter struct {
	*redis.Client
}

func NewFailureLimiter(client *redis.Client) auth.FailureLimiter {
	return &failureLimiter{Client: client}
}

func (l *failureLimiter) RetryAfter(ctx context.Context, key string, maxFailures int) (time.Duration, error) {
	redisKey := failuresKey(key)

	pipe := l.Client.Pipeline()
	count := pipe.Get(ctx, redisKey)
	ttl := pipe.PTTL(ctx, redisKey)

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return 0, err
	}

	failures, err := count.Int()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	if failures < maxFailures {
		return 0, nil
	}

	remaining, err := ttl.Result()
	if err != nil {
		return 0, err
	}

	return max(remaining, 0), nil
}

func (l *failureLimiter) RecordFailure(ctx context.Context, key string, window time.Duration) error {
	redisKey := failuresKey(key)

	pipe := l.Client.TxPipeline()
	pipe.Incr(ctx, redisKey)
	pipe.ExpireNX(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)

	return err
}

func (l *failureLimiter) Reset(ctx context.Context, key string) error {
	return l.Client.Del(ctx, failuresKey(key)).Err()
}

func failuresKey(key string) string {
	return "failures:" + hashValue(key)
}
