//go:build integration

package redis_test

import (
	"context"
	"encoding/json"
	"finance-app-go-demo/internal/application/auth"
	infraredis "finance-app-go-demo/internal/infrastructure/redis"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const testTTL = time.Hour

func TestSessionStorage_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	userID := uuid.New()

	before := time.Now()
	created, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	require.Equal(t, userID, created.UserID)
	require.WithinDuration(t, before.Add(testTTL), created.ExpiresAt, 5*time.Second)

	fetched, err := storage.GetSession(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, userID, fetched.UserID)
	require.True(t, created.ExpiresAt.Equal(fetched.ExpiresAt), "expiresAt must survive the round trip")
}

func TestSessionStorage_CreateGeneratesUniqueIDs(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	userID := uuid.New()

	first, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)
	second, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)

	require.NotEqual(t, first.ID, second.ID, "a user may hold several sessions")

	for _, s := range []*auth.Session{first, second} {
		fetched, err := storage.GetSession(ctx, s.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
	}
}

func TestSessionStorage_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	fetched, err := storage.GetSession(ctx, uuid.NewString())
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestSessionStorage_Delete(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	kept, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)
	deleted, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)

	require.NoError(t, storage.DeleteSession(ctx, deleted.ID))

	gone, err := storage.GetSession(ctx, deleted.ID)
	require.NoError(t, err)
	require.Nil(t, gone)

	still, err := storage.GetSession(ctx, kept.ID)
	require.NoError(t, err)
	require.NotNil(t, still, "deleting one session must not affect others")
}

func TestSessionStorage_Delete_Missing(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	require.NoError(t, storage.DeleteSession(ctx, uuid.NewString()))
}

func TestSessionStorage_StoredUnderPrefixedKeyWithTTL(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	created, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)

	raw, err := testClient.Get(ctx, "session:"+created.ID).Result()
	require.NoError(t, err, "session must live under the session: prefix")

	var decoded struct {
		UserID uuid.UUID `json:"userId"`
	}
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	require.Equal(t, created.UserID, decoded.UserID)

	ttl, err := testClient.TTL(ctx, "session:"+created.ID).Result()
	require.NoError(t, err)
	require.Greater(t, ttl, testTTL-time.Minute)
	require.LessOrEqual(t, ttl, testTTL)
}

func TestSessionStorage_ExpiresAfterTTL(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, 1*time.Second)

	created, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)

	fetched, err := storage.GetSession(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	require.Eventually(t, func() bool {
		s, err := storage.GetSession(ctx, created.ID)
		return err == nil && s == nil
	}, 5*time.Second, 100*time.Millisecond, "session must disappear once the TTL elapses")
}

func TestSessionStorage_Get_EmptyValue(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	id := uuid.NewString()

	require.NoError(t, testClient.Set(ctx, "session:"+id, "", testTTL).Err())

	fetched, err := storage.GetSession(ctx, id)
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestSessionStorage_Get_CorruptedValue(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	id := uuid.NewString()

	require.NoError(t, testClient.Set(ctx, "session:"+id, "{not json", testTTL).Err())

	fetched, err := storage.GetSession(ctx, id)
	require.Error(t, err)
	require.Nil(t, fetched)
}

func TestSessionStorage_DoesNotReadRecoveryKeys(t *testing.T) {
	ctx := context.Background()
	sessions := infraredis.NewSessionStorage(testClient, testTTL)
	recoveries := infraredis.NewRecoveryStorage(testClient, testTTL)

	recovery, err := recoveries.CreateRecovery(ctx, uuid.New())
	require.NoError(t, err)

	// a recovery code must never be usable as a session id
	session, err := sessions.GetSession(ctx, recovery.Code)
	require.NoError(t, err)
	require.Nil(t, session)
}

func TestSessionStorage_DeleteSessionsByUserID(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	userID := uuid.New()

	first, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)
	second, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)
	other, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)

	require.NoError(t, storage.DeleteSessionsByUserID(ctx, userID, ""))

	for _, id := range []string{first.ID, second.ID} {
		gone, err := storage.GetSession(ctx, id)
		require.NoError(t, err)
		require.Nil(t, gone, "every session of the user must be gone")
	}

	still, err := storage.GetSession(ctx, other.ID)
	require.NoError(t, err)
	require.NotNil(t, still, "other users' sessions must be untouched")

	exists, err := testClient.Exists(ctx, "user_sessions:"+userID.String()).Result()
	require.NoError(t, err)
	require.Zero(t, exists, "the index must be removed together with the sessions")
}

func TestSessionStorage_DeleteSessionsByUserID_NoSessions(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	require.NoError(t, storage.DeleteSessionsByUserID(ctx, uuid.New(), ""))
}

func TestSessionStorage_DeleteSession_RemovesItFromUserIndex(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	userID := uuid.New()

	kept, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)
	deleted, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)

	require.NoError(t, storage.DeleteSession(ctx, deleted.ID))

	members, err := testClient.SMembers(ctx, "user_sessions:"+userID.String()).Result()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{kept.ID}, members)
}

func TestSessionStorage_UserIndexExpiresWithSessions(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	userID := uuid.New()

	_, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)

	ttl, err := testClient.TTL(ctx, "user_sessions:"+userID.String()).Result()
	require.NoError(t, err)
	require.Greater(t, ttl, testTTL-time.Minute)
	require.LessOrEqual(t, ttl, testTTL)
}

func TestSessionStorage_DeleteSessionsByUserID_KeepsTheExcludedSession(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)
	userID := uuid.New()

	current, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)
	other, err := storage.CreateSession(ctx, userID)
	require.NoError(t, err)

	require.NoError(t, storage.DeleteSessionsByUserID(ctx, userID, current.ID))

	kept, err := storage.GetSession(ctx, current.ID)
	require.NoError(t, err)
	require.NotNil(t, kept)

	gone, err := storage.GetSession(ctx, other.ID)
	require.NoError(t, err)
	require.Nil(t, gone)

	members, err := testClient.SMembers(ctx, "user_sessions:"+userID.String()).Result()
	require.NoError(t, err)
	require.ElementsMatch(t, []string{current.ID}, members)

	require.NoError(t, storage.DeleteSessionsByUserID(ctx, userID, ""))
	gone, err = storage.GetSession(ctx, current.ID)
	require.NoError(t, err)
	require.Nil(t, gone, "a later call without an exception removes the rest")
}

func TestSessionStorage_ShortenSession(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	created, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)

	require.NoError(t, storage.ShortenSession(ctx, created.ID, 30*time.Second))

	ttl, err := testClient.TTL(ctx, "session:"+created.ID).Result()
	require.NoError(t, err)
	require.LessOrEqual(t, ttl, 30*time.Second)
	require.Greater(t, ttl, 20*time.Second)

	still, err := storage.GetSession(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, still, "the session keeps working during the grace period")
}

func TestSessionStorage_ShortenSession_NeverExtends(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, 10*time.Second)

	created, err := storage.CreateSession(ctx, uuid.New())
	require.NoError(t, err)

	require.NoError(t, storage.ShortenSession(ctx, created.ID, time.Hour))

	ttl, err := testClient.TTL(ctx, "session:"+created.ID).Result()
	require.NoError(t, err)
	require.LessOrEqual(t, ttl, 10*time.Second)
}

func TestSessionStorage_ShortenSession_Unknown(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewSessionStorage(testClient, testTTL)

	require.NoError(t, storage.ShortenSession(ctx, uuid.NewString(), time.Second))
}
