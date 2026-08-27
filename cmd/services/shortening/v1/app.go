package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/core/services/shortening"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmapping"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/snowflake"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	serviceCfg := &cfg.Services.Shortening
	logger := logging.New(cfg).With(slog.String("service", serviceCfg.Name))

	var cassandraCtx infra.Cassandra

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		return 1
	}
	inf.BindCassandra(cfg.Infrastructure.Cassandra, &cassandraCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	urlmappingRepo := urlmapping.NewCassandra(cassandraCtx.Session, cfg.Infrastructure.Cassandra.Keyspace, ports.NewCacheStub())

	idGenerator := snowflake.New("sa")
	proto := "https"
	if cfg.Env == config.EnvLocal {
		proto = "http"
	}
	shortURLPrefix := fmt.Sprintf("%s://%s/", proto, cfg.App.Domain)
	shorteningCfg := shortening.Config{
		ShortURLPrefix: shortURLPrefix,
		URLMappingRepo: urlmappingRepo,
		IdGenerator:    idGenerator,
	}
	service := shortening.New(shorteningCfg)

	readiness := func(ctx context.Context) error {
		if err := cassandraCtx.Ping(ctx); err != nil {
			return fmt.Errorf("'cassandraCtx.Ping' failed: %w", err)
		}
		return nil
	}

	handlers := NewHandlers(service)

	mux := web.NewMux(logger)
	mux.Use(web.Trace)
	mux.Use(web.Logging)
	mux.Use(web.Measure)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	mux.Post("/v1/shorten", handlers.shorten)

	app := web.New(serviceCfg, cfg, logger)
	if err := app.Run(mux, readiness); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
