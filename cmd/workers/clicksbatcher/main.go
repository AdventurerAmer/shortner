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
	"github.com/AdventurerAmer/shortner/internal/repos/analyticstat"
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

	cassandra, err := infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
	if err != nil {
		logger.Error("cassandra connection failed", "error", err)
		os.Exit(1)
	}
	defer infra.CloseCassandra(context.TODO(), cassandra)

	analyticStatRepo := analyticstat.NewCassandra(
		cassandra.Session,
		cfg.Infrastructure.Database.Keyspace,
		ports.NewCacheStub())

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

		// TODO: hardcoding timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// TODO: we have need manual ack here...
		if err := analyticStatRepo.Put(ctx, event.UUId, event.Aliases, event.Clicks); err != nil {
			logger.Info("analyticStatRepo.Put failed", "error", err)
		}
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		slog.Info("'consumer.Receive' failed", "error", err)
	}
}
