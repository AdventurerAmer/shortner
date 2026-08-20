package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/apps/worker"
	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	workerCfg := &cfg.Workers.Clicks
	logger := logging.New(cfg).With(slog.String("worker", workerCfg.Name))

	writer := infra.NewKafkaWriter(cfg.Infrastructure.Kafka, domain.ClicksBatchTopic)
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("'producer.Close' failed", "error", err)
		}
	}()

	producer := brokers.NewKafkaProducer(writer)

	orch := goorch.New(context.Background())
	defer orch.CancelAndWait()

	eventProducer, err := ports.NewEventProducer(producer, orch)
	if err != nil {
		logger.Error("create event producer failed", "error", err)
		return 1
	}

	bucketCount := 256
	bucketCapacity := 256
	batchSize := 1024
	collector := newCollector(bucketCount, bucketCapacity, batchSize, orch, eventProducer)

	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.Topic(workerCfg.Topic), workerCfg.Group)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)
	app := worker.New(&cfg.Workers.ClicksBatcher, cfg, consumer, logger)

	h := func(hctx context.Context, msg ports.ConsumerMessage) error {
		dctx, span := telemetry.NewSpan(hctx, "Clicks Batcher Worker: Handle", telemetry.StrAttr("Key", msg.Key))
		defer span.End()

		logger := logging.Get(dctx)

		logger.Info("recived event", "status", "started", "key", msg.Key)
		defer logger.Info("recived event", "status", "ended", "key", msg.Key)

		var event domain.ClickEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return fmt.Errorf("'json.Unmarshal' failed: %w", err)
		}

		alias := event.Alias
		collector.inc(alias)

		return nil
	}
	if err := app.Run(h); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
