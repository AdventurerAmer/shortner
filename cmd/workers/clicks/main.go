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
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticclicks"
	"github.com/AdventurerAmer/shortner/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	groupId := "clicks"
	logger := logging.New(cfg).With(slog.String("service", groupId))

	clickHouse, err := infra.ConnectClickHouse(context.TODO(), &cfg.Infrastructure.ClickHouse)
	if err != nil {
		logger.Error("clickhouse connection failed", "error", err)
		os.Exit(1)
	}
	defer infra.CloseClickHouse(context.TODO(), clickHouse)

	logger.Info("Connected to ClickHouse")

	analyticClicksRepo := analyticclicks.NewClickHouse(
		cfg.Infrastructure.ClickHouse.Database, clickHouse.Conn, ports.NewCacheStub(), time.Second)

	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.ClicksBatchTopic, groupId)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)

	h := func(ctx context.Context, key string, data []byte) error {
		logger.Info("recived event", "status", "started", "key", key)
		defer logger.Info("recived event", "status", "ended", "key", key)

		var event domain.ClicksBatchEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("'json.Unmarshal' failed: %w", err)
		}

		// TODO: hardcoding timeout
		dctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// TODO: we have need manual ack here...
		if err := analyticClicksRepo.Put(dctx, event.UUIds, event.Aliases, event.Clicks); err != nil {
			return fmt.Errorf("'analyticClicksRepo.Put' failed: %w", err)
		}

		return nil
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		logger.Error("'consumer.Receive' failed", "error", err)
	}
}
