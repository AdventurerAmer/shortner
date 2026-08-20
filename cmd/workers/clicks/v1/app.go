package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/shortner/apps/worker"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticclicks"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	workerCfg := &cfg.Workers.ClicksBatcher
	logger := logging.New(cfg).With(slog.String("worker", workerCfg.Name))

	var clickHouseCtx infra.ClickHouse

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		return 1
	}
	inf.BindClickHouse(cfg.Infrastructure.ClickHouse, &clickHouseCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	analyticClicksRepo := analyticclicks.NewClickHouse(
		cfg.Infrastructure.ClickHouse.Database, clickHouseCtx.Conn, ports.NewCacheStub(), time.Second)

	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.Topic(workerCfg.Topic), workerCfg.Group)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)
	app := worker.New(&cfg.Workers.Clicks, cfg, consumer, logger)

	h := func(ctx context.Context, msg ports.ConsumerMessage) error {
		dctx, span := telemetry.NewSpan(ctx, "Clicks Worker: Handle",
			telemetry.StrAttr("key", msg.Key))
		defer span.End()

		logger := logging.Get(dctx)

		logger.Info("recived event", "status", "started", "key", msg.Key)
		defer logger.Info("recived event", "status", "ended", "key", msg.Key)

		var event domain.ClicksBatchEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return fmt.Errorf("'json.Unmarshal' failed: %w", err)
		}

		dctx, cancel := context.WithTimeout(ctx, domain.SendEventDefaultTimeout)
		defer cancel()

		if err := analyticClicksRepo.Put(dctx, event.UUIds, event.Aliases, event.Clicks); err != nil {
			return fmt.Errorf("'analyticClicksRepo.Put' failed: %w", err)
		}

		return nil
	}
	if err := app.Run(h); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
