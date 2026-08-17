package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/redis/go-redis/v9"
)

type redisConfigWrapper struct {
	config.RedisConfig
}

func (cfg *redisConfigWrapper) Connect(ctx context.Context) (Disconnecter, error) {
	database := 0
	if cfg.Database != nil {
		database = *cfg.Database
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	opts := &redis.Options{
		Addr:     addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       database,
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("'client.Ping' failed: %w", err)
	}
	return &Redis{Client: client}, nil
}

type Redis struct {
	Client *redis.Client
}

func (r *Redis) Disconnect(ctx context.Context) error {
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
