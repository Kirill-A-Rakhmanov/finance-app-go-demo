//go:build integration

package redis_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"finance-app-go-demo/internal/application/auth"
	infraredis "finance-app-go-demo/internal/infrastructure/redis"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newPending(email string) auth.PendingRegistration {
	return auth.PendingRegistration{
		FirstName: "Jane", LastName: "Doe", Email: email, PasswordHash: "hashed-password",
	}
}

func uniqueEmail() string {
	return uuid.NewString() + "@example.com"
}

func TestRegistrationStorage_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)
	email := uniqueEmail()

	token, err := storage.Create(ctx, newPending(email))
	require.NoError(t, err)
	require.NotEmpty(t, token)

	got, err := storage.Get(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, newPending(email), *got)
}

func TestRegistrationStorage_Get_UnknownToken(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)

	got, err := storage.Get(ctx, "no-such-token")
	require.NoError(t, err)
	require.Nil(t, got)

	got, err = storage.Get(ctx, "")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestRegistrationStorage_TokensAreUnique(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)

	first, err := storage.Create(ctx, newPending(uniqueEmail()))
	require.NoError(t, err)
	second, err := storage.Create(ctx, newPending(uniqueEmail()))
	require.NoError(t, err)

	require.NotEqual(t, first, second)
	require.GreaterOrEqual(t, len(first), 43, "256 random bits, base64url encoded")
}

func TestRegistrationStorage_NewRegistrationReplacesThePreviousOne(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)
	email := uniqueEmail()

	oldToken, err := storage.Create(ctx, newPending(email))
	require.NoError(t, err)

	replacement := newPending(email)
	replacement.FirstName = "Replaced"
	newToken, err := storage.Create(ctx, replacement)
	require.NoError(t, err)

	gone, err := storage.Get(ctx, oldToken)
	require.NoError(t, err)
	require.Nil(t, gone, "the old link must stop working")

	got, err := storage.Get(ctx, newToken)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "Replaced", got.FirstName)
}

func TestRegistrationStorage_EmailsDifferingInCaseShareOneEntry(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)
	email := uniqueEmail()

	oldToken, err := storage.Create(ctx, newPending(email))
	require.NoError(t, err)
	_, err = storage.Create(ctx, newPending(strings.ToUpper(email)))
	require.NoError(t, err)

	gone, err := storage.Get(ctx, oldToken)
	require.NoError(t, err)
	require.Nil(t, gone)
}

func TestRegistrationStorage_Delete(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)

	deleted, err := storage.Create(ctx, newPending(uniqueEmail()))
	require.NoError(t, err)
	kept, err := storage.Create(ctx, newPending(uniqueEmail()))
	require.NoError(t, err)

	require.NoError(t, storage.Delete(ctx, deleted))

	gone, err := storage.Get(ctx, deleted)
	require.NoError(t, err)
	require.Nil(t, gone)

	still, err := storage.Get(ctx, kept)
	require.NoError(t, err)
	require.NotNil(t, still, "deleting one registration must not affect others")
}

func TestRegistrationStorage_Delete_Unknown(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)

	require.NoError(t, storage.Delete(ctx, "no-such-token"))
}

func TestRegistrationStorage_ExpiresAfterTTL(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, time.Second)

	token, err := storage.Create(ctx, newPending(uniqueEmail()))
	require.NoError(t, err)

	got, err := storage.Get(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Eventually(t, func() bool {
		g, err := storage.Get(ctx, token)
		return err == nil && g == nil
	}, 5*time.Second, 100*time.Millisecond, "an unconfirmed registration must expire on its own")
}

func TestRegistrationStorage_TokenIsNotStoredInPlainText(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)

	token, err := storage.Create(ctx, newPending(uniqueEmail()))
	require.NoError(t, err)

	keys, err := testClient.Keys(ctx, "registration:*"+token+"*").Result()
	require.NoError(t, err)
	require.Empty(t, keys, "only a hash of the token may appear in Redis keys")
}

func TestRegistrationStorage_AcquireCooldown(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)
	email := uniqueEmail()

	first, err := storage.AcquireCooldown(ctx, email, time.Minute)
	require.NoError(t, err)
	require.True(t, first)

	again, err := storage.AcquireCooldown(ctx, email, time.Minute)
	require.NoError(t, err)
	require.False(t, again, "a second attempt within the cooldown is refused")

	differentCase, err := storage.AcquireCooldown(ctx, strings.ToUpper(email), time.Minute)
	require.NoError(t, err)
	require.False(t, differentCase, "the cooldown ignores letter case")

	other, err := storage.AcquireCooldown(ctx, uniqueEmail(), time.Minute)
	require.NoError(t, err)
	require.True(t, other, "other emails are unaffected")
}

func TestRegistrationStorage_AcquireCooldown_ExpiresAfterCooldown(t *testing.T) {
	ctx := context.Background()
	storage := infraredis.NewRegistrationStorage(testClient, testTTL)
	email := uniqueEmail()

	ok, err := storage.AcquireCooldown(ctx, email, time.Second)
	require.NoError(t, err)
	require.True(t, ok)

	require.Eventually(t, func() bool {
		ok, err := storage.AcquireCooldown(ctx, email, time.Second)
		return err == nil && ok
	}, 5*time.Second, 100*time.Millisecond)
}
