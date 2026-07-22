package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/caches"
	"github.com/AdventurerAmer/shortner/internal/core/services/analytics"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticclicks"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	serviceCfg := &cfg.Services.Analytics
	logger := logging.New(cfg).With(slog.String("service", serviceCfg.Name))

	cassandra, err := infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Cassandra)
	if err != nil {
		logger.Error("cassandra connection failed", "error", err)
		return 1
	}
	defer infra.CloseCassandra(context.TODO(), cassandra)

	redisCtx, err := infra.ConnectToRedis(context.TODO(), &cfg.Infrastructure.RedisAnalytics)
	if err != nil {
		logger.Error("redis connection failed", "error", err)
		return 1
	}
	defer infra.CloseRedis(context.TODO(), redisCtx)

	clickHouse, err := infra.ConnectClickHouse(context.TODO(), &cfg.Infrastructure.ClickHouse)
	if err != nil {
		logger.Error("clickhouse connection failed", "error", err)
		os.Exit(1)
	}
	defer infra.CloseClickHouse(context.TODO(), clickHouse)

	redisCache := caches.NewRedis(redisCtx.Client)

	analyticClicksRepo := analyticclicks.NewClickHouse(
		cfg.Infrastructure.ClickHouse.Database, clickHouse.Conn, redisCache, time.Second)

	analyticsCfg := analytics.Config{
		AnalyticStatRepo: analyticClicksRepo,
	}
	service := analytics.New(analyticsCfg)

	mux := web.NewMux(logger)

	mux.Use(web.RequestId)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	handlers := newHandlers(service)

	mux.Get("/v1/analytics/{alias}", handlers.get)

	app := web.New(serviceCfg, logger)
	app.Run(mux)

	return 0
}
