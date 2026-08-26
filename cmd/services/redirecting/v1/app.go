package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/caches"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/core/services/redirecting"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmapping"
	"github.com/AdventurerAmer/shortner/logging"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	serviceCfg := &cfg.Services.Redirecting
	logger := logging.New(cfg).With(slog.String("service", serviceCfg.Name))

	var (
		redisCtx     infra.Redis
		cassandraCtx infra.Cassandra
	)

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		return 1
	}
	inf.BindRedis(cfg.Infrastructure.Redis, &redisCtx)
	inf.BindCassandra(cfg.Infrastructure.Cassandra, &cassandraCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	redisCache := caches.NewRedis(redisCtx.Client)

	URLMappingRepo := urlmapping.NewCassandra(
		cassandraCtx.Session,
		cfg.Infrastructure.Cassandra.Keyspace,
		redisCache)

	redirectingCfg := redirecting.Config{
		URLMappingRepo: URLMappingRepo,
	}
	service := redirecting.New(redirectingCfg)

	orch := goorch.New(context.Background())
	defer orch.CancelAndWait()

	writer := infra.NewKafkaWriter(cfg.Infrastructure.Kafka, domain.ClicksTopic)
	defer func() {
		_ = writer.Close()
	}()

	producer := brokers.NewKafkaProducer(writer)

	eventProducer, err := ports.NewEventProducer(producer, orch)
	if err != nil {
		logger.Error("create event producer failed", "error", err)
		return 1
	}

	handlers := newHandlers(service, eventProducer)

	mux := web.NewMux(logger)

	mux.Use(web.Trace)
	mux.Use(web.Measure)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	mux.Get("/v1/redirect/{alias}", handlers.redirect)

	app := web.New(serviceCfg, cfg, logger)
	if err := app.Run(mux); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
