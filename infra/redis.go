package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/redis/go-redis/v9"
)

func ConnectToRedis(ctx context.Context, cfg *config.RedisConfig) (config.Redis, error) {
	database := 0
	if cfg.Database != nil {
		database = *cfg.Database
	}
	opts := &redis.Options{
		Addr:     cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       database,
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return config.Redis{}, fmt.Errorf("'client.Ping' failed: %w", err)
	}
	return config.Redis{Client: client}, nil
}

func CloseRedis(ctx context.Context, r config.Redis) error {
	errCh := make(chan error)
	go func() {
		if err := r.Client.Close(); err != nil {
			errCh <- fmt.Errorf("'Client.Close' failed: %w", err)
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
