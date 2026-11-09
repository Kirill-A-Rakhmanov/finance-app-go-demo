package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `envconfig:"REDIS_HOST" required:"true" default:"localhost"`
	Port     string `envconfig:"REDIS_PORT" required:"true" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" required:"true"`
}

func New(c Config) (*redis.Client, error) {
	addr := c.Host + ":" + c.Port

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: c.Password,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
