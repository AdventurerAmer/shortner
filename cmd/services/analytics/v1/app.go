package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/health"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/caches"
	"github.com/AdventurerAmer/shortner/internal/core/services/analytics"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticclicks"
	"github.com/AdventurerAmer/shortner/logging"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	serviceCfg := &cfg.Services.Analytics
	logger := logging.New(cfg).With(slog.String("service", serviceCfg.Name))

	var (
		redisCtx      infra.Redis
		clickHouseCtx infra.ClickHouse
	)

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		return 1
	}
	inf.BindRedis(cfg.Infrastructure.RedisAnalytics, &redisCtx)
	inf.BindClickHouse(cfg.Infrastructure.ClickHouse, &clickHouseCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		return 1
	}
	defer inf.Shutdown(context.Background())

	redisCache := caches.NewRedis(redisCtx.Client)

	ttl := cfg.Constants.AnalyticClicksCacheTTL
	logger.Info("cfg.Constants.AnalyticClicksCacheTTL", "ttl", ttl)
	analyticClicksRepo := analyticclicks.NewClickHouse(cfg.Infrastructure.ClickHouse.Database, clickHouseCtx.Conn, redisCache, ttl)

	analyticsCfg := analytics.Config{
		AnalyticClicksRepo: analyticClicksRepo,
	}
	service := analytics.New(analyticsCfg)

	readiness := func(ctx context.Context, checks health.Checks) error {
		if err := clickHouseCtx.Conn.Ping(ctx); err != nil {
			checks["clickhouse"] = err.Error()
			return fmt.Errorf("'clickHouseCtx.Conn.Ping' failed: %w", err)
		}
		checks["clickhouse"] = "up"

		if _, err := redisCtx.Client.Ping(ctx).Result(); err != nil {
			checks["redis"] = err.Error()
			return fmt.Errorf("'redisCtx.Client.Ping' failed: %w", err)
		}

		checks["redis"] = "up"

		return nil
	}

	mux := web.NewMux(logger)

	mux.Use(web.Trace)
	mux.Use(web.Measure)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	handlers := newHandlers(service)

	mux.Get("/v1/analytics/{alias}", handlers.get)

	app := web.New(serviceCfg, cfg, logger)
	if err := app.Run(mux, readiness); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
