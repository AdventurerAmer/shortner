package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/config"
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

	analyticClicksRepo := analyticclicks.NewClickHouse(
		cfg.Infrastructure.ClickHouse.Database, clickHouseCtx.Conn, redisCache, time.Second)

	analyticsCfg := analytics.Config{
		AnalyticStatRepo: analyticClicksRepo,
	}
	service := analytics.New(analyticsCfg)

	mux := web.NewMux(logger)

	mux.Use(web.Trace(serviceCfg.Name))
	mux.Use(web.CorrelationId)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	handlers := newHandlers(service)

	mux.Get("/v1/analytics/{alias}", handlers.get)

	app := web.New(cfg, serviceCfg, logger)
	if err := app.Run(mux); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		return 1
	}

	return 0
}
