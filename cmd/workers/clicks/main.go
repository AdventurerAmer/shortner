package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/caches"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg).With(slog.String("service", "clicks"))

	redisCtx, err := infra.ConnectToRedis(context.TODO(), &cfg.Infrastructure.RedisAnalytics)
	if err != nil {
		logger.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer infra.CloseRedis(context.TODO(), redisCtx)

	cache := caches.NewRedis(redisCtx.Client)

	groupId := "clicks"
	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.ClicksTopic, groupId)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)
	h := func(key string, data []byte) {
		logger.Info("recived event", "key", key)

		var event domain.ClickEvent
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("'json.Unmarshal' failed", "key", key, "error", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		alias := event.Alias
		if err := cache.Inc(ctx, alias); err != nil {
			logger.Error("'cache.Inc' failed", "key", key, "event", event, "error", err)
			return
		}
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		slog.Info("'consumer.Receive' failed", "error", err)
	}
}
