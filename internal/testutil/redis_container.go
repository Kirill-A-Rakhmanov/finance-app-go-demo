//go:build integration

package testutil

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartRedis(ctx context.Context) (*redis.Client, func(), error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections"),
		},
		Started: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("start redis container: %w", err)
	}

	terminate := func() { _ = container.Terminate(ctx) }

	host, err := container.Host(ctx)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("redis host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("redis port: %w", err)
	}

	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", host, port.Port())})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		terminate()
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, func() {
		_ = client.Close()
		terminate()
	}, nil
}
