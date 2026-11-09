//go:build integration

package redis_test

import (
	"context"
	"encoding/json"
	infraredis "finance-app-go-demo/internal/infrastructure/redis"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRecoveryStorage_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)
	userID := uuid.New()

	before := time.Now()
	created, err := storage.CreateRecovery(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, created.Code)
	require.Equal(t, userID, created.UserID)
	require.WithinDuration(t, before.Add(testTTL), created.ExpiresAt, 5*time.Second)

	fetched, err := storage.GetRecovery(ctx, created.Code)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, created.Code, fetched.Code)
	require.Equal(t, userID, fetched.UserID)
	require.True(t, created.ExpiresAt.Equal(fetched.ExpiresAt), "expiresAt must survive the round trip")
}

func TestRecoveryStorage_CreateGeneratesUniqueCodes(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)
	userID := uuid.New()

	first, err := storage.CreateRecovery(ctx, userID)
	require.NoError(t, err)
	second, err := storage.CreateRecovery(ctx, userID)
	require.NoError(t, err)

	require.NotEqual(t, first.Code, second.Code)
}

func TestRecoveryStorage_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)

	fetched, err := storage.GetRecovery(ctx, uuid.NewString())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestRecoveryStorage_Delete(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)

	kept, err := storage.CreateRecovery(ctx, uuid.New())
	require.NoError(t, err)
	deleted, err := storage.CreateRecovery(ctx, uuid.New())
	require.NoError(t, err)

	require.NoError(t, storage.DeleteRecovery(ctx, deleted.Code))

	gone, err := storage.GetRecovery(ctx, deleted.Code)
	require.NoError(t, err)
	require.Nil(t, gone, "a used recovery code must not be reusable")

	still, err := storage.GetRecovery(ctx, kept.Code)
	require.NoError(t, err)
	require.NotNil(t, still)
}

func TestRecoveryStorage_Delete_Missing(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)

	require.NoError(t, storage.DeleteRecovery(ctx, uuid.NewString()))
}

func TestRecoveryStorage_StoredUnderPrefixedKeyWithTTL(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)

	created, err := storage.CreateRecovery(ctx, uuid.New())
	require.NoError(t, err)

	raw, err := testClient.Get(ctx, "recovery:"+created.Code).Result()
	require.NoError(t, err, "recovery must live under the recovery: prefix")

	var decoded struct {
		UserID uuid.UUID `json:"userId"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	require.Equal(t, created.UserID, decoded.UserID)

	ttl, err := testClient.TTL(ctx, "recovery:"+created.Code).Result()
	require.NoError(t, err)
	require.Greater(t, ttl, testTTL-time.Minute)
	require.LessOrEqual(t, ttl, testTTL)
}

func TestRecoveryStorage_ExpiresAfterTTL(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, 1*time.Second)

	created, err := storage.CreateRecovery(ctx, uuid.New())
	require.NoError(t, err)

	fetched, err := storage.GetRecovery(ctx, created.Code)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	require.Eventually(t, func() bool {
		r, err := storage.GetRecovery(ctx, created.Code)
		return err == nil && r == nil
	}, 5*time.Second, 100*time.Millisecond, "recovery must disappear once the TTL elapses")
}

func TestRecoveryStorage_Get_EmptyValue(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)
	code := uuid.NewString()

	require.NoError(t, testClient.Set(ctx, "recovery:"+code, "", testTTL).Err())

	fetched, err := storage.GetRecovery(ctx, code)
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestRecoveryStorage_Get_CorruptedValue(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)
	code := uuid.NewString()

	require.NoError(t, testClient.Set(ctx, "recovery:"+code, "{not json", testTTL).Err())

	fetched, err := storage.GetRecovery(ctx, code)
	require.Error(t, err)
	require.Nil(t, fetched)
}

func TestRecoveryStorage_AcquireCooldown(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)
	email := uuid.NewString() + "@example.com"

	first, err := storage.AcquireCooldown(ctx, email, time.Minute)
	require.NoError(t, err)
	require.True(t, first)

	again, err := storage.AcquireCooldown(ctx, email, time.Minute)
	require.NoError(t, err)
	require.False(t, again, "a second request within the cooldown is refused")

	differentCase, err := storage.AcquireCooldown(ctx, "  "+strings.ToUpper(email), time.Minute)
	require.NoError(t, err)
	require.False(t, differentCase, "the cooldown ignores letter case and surrounding spaces")

	other, err := storage.AcquireCooldown(ctx, uuid.NewString()+"@example.com", time.Minute)
	require.NoError(t, err)
	require.True(t, other)

	keys, err := testClient.Keys(ctx, "recovery-cooldown:*"+email+"*").Result()
	require.NoError(t, err)
	require.Empty(t, keys, "the email must not appear in Redis keys")
}

func TestRecoveryStorage_AcquireCooldown_ExpiresAfterCooldown(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRecoveryStorage(testClient, testTTL)
	email := uuid.NewString() + "@example.com"

	ok, err := storage.AcquireCooldown(ctx, email, time.Second)
	require.NoError(t, err)
	require.True(t, ok)

	require.Eventually(t, func() bool {
		ok, err := storage.AcquireCooldown(ctx, email, time.Second)
		return err == nil && ok
	}, 5*time.Second, 100*time.Millisecond)
}
