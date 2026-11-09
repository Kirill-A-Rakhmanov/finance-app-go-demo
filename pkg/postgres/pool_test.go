package postgres

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurePool_CapsOpenConnections(t *testing.T) {
	db, err := sql.Open("postgres", Config{Host: "localhost", Port: "5432", User: "u", Password: "p", DBName: "d"}.URL())
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.Zero(t, db.Stats().MaxOpenConnections, "an unconfigured pool is unlimited")

	configurePool(db)

	assert.Equal(t, MaxOpenConns, db.Stats().MaxOpenConnections)
}

func TestPoolLimits_AreConsistent(t *testing.T) {
	assert.Positive(t, MaxOpenConns)
	assert.LessOrEqual(t, MaxIdleConns, MaxOpenConns, "idle connections are a subset of the open ones")
	assert.Less(t, MaxOpenConns, 100, "must stay under the PostgreSQL default max_connections")
	assert.Greater(t, ConnMaxLifetime, ConnMaxIdleTime)
}
