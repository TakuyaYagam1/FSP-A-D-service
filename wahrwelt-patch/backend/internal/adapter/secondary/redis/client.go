package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/config"
	"github.com/redis/go-redis/v9"
)

const namespace = "wahrwelt"

type Store struct{ client *redis.Client }

func New(ctx context.Context, cfg config.Config) (*Store, func(), error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	cleanup := func() { _ = client.Close() }
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(connectCtx).Err(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("connect to Redis: %w", err)
	}
	return &Store{client: client}, cleanup, nil
}
