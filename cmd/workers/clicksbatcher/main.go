package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	groupId := "clicks-batcher"
	logger := logging.New(cfg).With(slog.String("service", groupId))

	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.ClicksBatchTopic, groupId)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)

	h := func(key string, data []byte) {
		logger.Info("recived event", "status", "started", "key", key)
		defer logger.Info("recived event", "status", "ended", "key", key)

		var event domain.ClicksBatchEvent
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("'json.Unmarshal' failed", "key", key, "error", err)
			return
		}
		slog.Info("processing batched clicks", "event", event)
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		slog.Info("'consumer.Receive' failed", "error", err)
	}
}
