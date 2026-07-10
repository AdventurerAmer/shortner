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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	groupId := "clicks-consumers"
	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.ClicksTopic, groupId)
	defer func() {
		_ = reader.Close()
	}()

	consumer := brokers.NewKafkaConsumer(reader)
	h := func(key string, data []byte) {
		slog.Info("recived event", "key", key)
		var event domain.ClickEvent
		if err := json.Unmarshal(data, &event); err != nil {
			// TODO: handle error here...
			return
		}
		slog.Info("recived event", "event", event)
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		slog.Info("'consumer.Receive' failed", "error", err)
	}
}
