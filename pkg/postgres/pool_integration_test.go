//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestConfigurePool_UnderLoad(t *testing.T) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:17",
		tcpostgres.WithDatabase("pool-test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	configurePool(db)

	var peak atomic.Int64
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				open := int64(db.Stats().OpenConnections)
				for {
					current := peak.Load()
					if open <= current || peak.CompareAndSwap(current, open) {
						break
					}
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < MaxOpenConns*3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := db.ExecContext(ctx, "SELECT pg_sleep(0.3)")
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	close(done)

	assert.LessOrEqual(t, int(peak.Load()), MaxOpenConns, "the pool never opens more connections than the limit")
	assert.Equal(t, MaxOpenConns, int(peak.Load()), "the load is big enough to reach the limit")
	assert.LessOrEqual(t, db.Stats().Idle, MaxIdleConns, "connections above the idle limit are closed")
}
