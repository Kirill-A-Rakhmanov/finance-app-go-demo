//go:build integration

package redis_test

import (
	"context"
	"finance-app-go-demo/internal/testutil"
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	client, cleanup, err := testutil.StartRedis(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start redis:", err)
		os.Exit(1)
	}
	testClient = client

	code := m.Run()

	cleanup()
	os.Exit(code)
}
