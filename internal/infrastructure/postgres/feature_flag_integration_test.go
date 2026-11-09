//go:build integration

package postgres_test

import (
	"context"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/infrastructure/postgres"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createTestFeatureFlag(t *testing.T, ctx context.Context, enabled bool) *model.FeatureFlag {
	t.Helper()

	created, err := postgres.NewFeatureFlagRepository(testClient).Create(
		ctx,
		model.NewFeatureFlag("flag-"+uuid.NewString(), "description", enabled),
	)
	require.NoError(t, err)

	return created
}

func TestFeatureFlagRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)

	created, err := repo.Create(ctx, model.NewFeatureFlag("registration-"+uuid.NewString(), "Open registration", true))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	byID, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, byID)
	require.Equal(t, created.Key, byID.Key)
	require.Equal(t, "Open registration", byID.Description)
	require.True(t, byID.Enabled)

	byKey, err := repo.GetByKey(ctx, created.Key)
	require.NoError(t, err)
	require.NotNil(t, byKey)
	require.Equal(t, created.ID, byKey.ID)
}

func TestFeatureFlagRepository_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)

	byID, err := repo.GetByID(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, byID)

	byKey, err := repo.GetByKey(ctx, "missing-"+uuid.NewString())
	require.NoError(t, err)
	require.Nil(t, byKey)
}

func TestFeatureFlagRepository_KeyIsUnique(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)
	existing := createTestFeatureFlag(t, ctx, false)

	_, err := repo.Create(ctx, model.NewFeatureFlag(existing.Key, "", false))

	require.Error(t, err)
}

func TestFeatureFlagRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)
	flag := createTestFeatureFlag(t, ctx, false)

	flag.Enabled = true
	flag.Description = "changed"

	updated, err := repo.Update(ctx, flag)
	require.NoError(t, err)
	require.True(t, updated.Enabled)
	require.Equal(t, "changed", updated.Description)

	fetched, err := repo.GetByID(ctx, flag.ID)
	require.NoError(t, err)
	require.True(t, fetched.Enabled)
	require.True(t, fetched.UpdatedAt.After(flag.UpdatedAt) || fetched.UpdatedAt.Equal(flag.UpdatedAt))
}

func TestFeatureFlagRepository_GetAll_OrderedByKey(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)
	prefix := "order-" + uuid.NewString() + "-"

	for _, suffix := range []string{"b", "c", "a"} {
		_, err := repo.Create(ctx, model.NewFeatureFlag(prefix+suffix, "", false))
		require.NoError(t, err)
	}

	all, err := repo.GetAll(ctx)
	require.NoError(t, err)

	var keys []string
	for _, flag := range all {
		if len(flag.Key) > len(prefix) && flag.Key[:len(prefix)] == prefix {
			keys = append(keys, flag.Key)
		}
	}

	require.Equal(t, []string{prefix + "a", prefix + "b", prefix + "c"}, keys)
}

func TestFeatureFlagRepository_DeleteByID_RemovesOverrides(t *testing.T) {
	ctx := context.Background()
	flags := postgres.NewFeatureFlagRepository(testClient)
	overrides := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	flag := createTestFeatureFlag(t, ctx, false)

	_, err := overrides.Create(ctx, model.NewFeatureFlagOverride(flag.ID, user.ID, true))
	require.NoError(t, err)

	require.NoError(t, flags.DeleteByID(ctx, flag.ID))

	fetched, err := flags.GetByID(ctx, flag.ID)
	require.NoError(t, err)
	require.Nil(t, fetched)

	remaining, err := overrides.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Empty(t, remaining, "overrides of a deleted flag must be deleted with it")
}

func TestFeatureFlagOverrideRepository_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	flag := createTestFeatureFlag(t, ctx, false)

	created, err := repo.Create(ctx, model.NewFeatureFlagOverride(flag.ID, user.ID, true))
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, created.ID)

	fetched, err := repo.GetByFeatureFlagIDAndUserID(ctx, flag.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, created.ID, fetched.ID)
	require.True(t, fetched.Enabled)
}

func TestFeatureFlagOverrideRepository_GetByFeatureFlagIDAndUserID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	flag := createTestFeatureFlag(t, ctx, false)

	_, err := repo.Create(ctx, model.NewFeatureFlagOverride(flag.ID, user.ID, true))
	require.NoError(t, err)

	foreign, err := repo.GetByFeatureFlagIDAndUserID(ctx, flag.ID, other.ID)
	require.NoError(t, err)
	require.Nil(t, foreign, "an override is visible only for its own user")
}

func TestFeatureFlagOverrideRepository_OnePerFlagAndUser(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	flag := createTestFeatureFlag(t, ctx, false)

	_, err := repo.Create(ctx, model.NewFeatureFlagOverride(flag.ID, user.ID, true))
	require.NoError(t, err)

	_, err = repo.Create(ctx, model.NewFeatureFlagOverride(flag.ID, user.ID, false))
	require.Error(t, err)
}

func TestFeatureFlagOverrideRepository_RequiresExistingFlagAndUser(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	flag := createTestFeatureFlag(t, ctx, false)

	_, err := repo.Create(ctx, model.NewFeatureFlagOverride(uuid.New(), user.ID, true))
	require.Error(t, err)

	_, err = repo.Create(ctx, model.NewFeatureFlagOverride(flag.ID, uuid.New(), true))
	require.Error(t, err)
}

func TestFeatureFlagOverrideRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	first := createTestFeatureFlag(t, ctx, false)
	second := createTestFeatureFlag(t, ctx, true)

	_, err := repo.Create(ctx, model.NewFeatureFlagOverride(first.ID, user.ID, true))
	require.NoError(t, err)
	_, err = repo.Create(ctx, model.NewFeatureFlagOverride(second.ID, user.ID, false))
	require.NoError(t, err)
	_, err = repo.Create(ctx, model.NewFeatureFlagOverride(first.ID, other.ID, true))
	require.NoError(t, err)

	result, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Len(t, result, 2)

	enabled := map[uuid.UUID]bool{}
	for _, o := range result {
		require.Equal(t, user.ID, o.UserID)
		enabled[o.FeatureFlagID] = o.Enabled
	}
	require.Equal(t, map[uuid.UUID]bool{first.ID: true, second.ID: false}, enabled)
}

func TestFeatureFlagOverrideRepository_UpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	flag := createTestFeatureFlag(t, ctx, false)

	created, err := repo.Create(ctx, model.NewFeatureFlagOverride(flag.ID, user.ID, true))
	require.NoError(t, err)

	created.Enabled = false

	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	require.False(t, updated.Enabled)

	require.NoError(t, repo.DeleteByID(ctx, created.ID))

	fetched, err := repo.GetByFeatureFlagIDAndUserID(ctx, flag.ID, user.ID)
	require.NoError(t, err)
	require.Nil(t, fetched)
}

func TestFeatureFlagRepository_IsEnabled(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)
	user := createTestUser(t, ctx)
	enabled := createTestFeatureFlag(t, ctx, true)
	disabled := createTestFeatureFlag(t, ctx, false)

	_, err := postgres.NewFeatureFlagOverrideRepository(testClient).
		Create(ctx, model.NewFeatureFlagOverride(disabled.ID, user.ID, true))
	require.NoError(t, err)

	on, err := repo.IsEnabled(ctx, enabled.Key)
	require.NoError(t, err)
	require.True(t, on)

	off, err := repo.IsEnabled(ctx, disabled.Key)
	require.NoError(t, err)
	require.False(t, off, "the global value ignores overrides")

	unknown, err := repo.IsEnabled(ctx, "missing-"+uuid.NewString())
	require.NoError(t, err)
	require.False(t, unknown, "an unknown flag is disabled")
}

func TestFeatureFlagRepository_IsEnabledForUser(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewFeatureFlagRepository(testClient)
	overrides := postgres.NewFeatureFlagOverrideRepository(testClient)
	user := createTestUser(t, ctx)
	other := createTestUser(t, ctx)
	enabled := createTestFeatureFlag(t, ctx, true)
	disabled := createTestFeatureFlag(t, ctx, false)

	_, err := overrides.Create(ctx, model.NewFeatureFlagOverride(enabled.ID, user.ID, false))
	require.NoError(t, err)
	_, err = overrides.Create(ctx, model.NewFeatureFlagOverride(disabled.ID, user.ID, true))
	require.NoError(t, err)

	tests := []struct {
		name   string
		key    string
		userID uuid.UUID
		want   bool
	}{
		{"override switches an enabled flag off", enabled.Key, user.ID, false},
		{"override switches a disabled flag on", disabled.Key, user.ID, true},
		{"another user gets the global value of an enabled flag", enabled.Key, other.ID, true},
		{"another user gets the global value of a disabled flag", disabled.Key, other.ID, false},
		{"unknown flag is disabled", "missing-" + uuid.NewString(), user.ID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.IsEnabledForUser(ctx, tt.key, tt.userID)

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
