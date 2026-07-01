package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/caches"
	"github.com/AdventurerAmer/shortner/internal/core/services/analytics"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticrepo"
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

	cassandra, err := infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
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

	redisCache := caches.NewRedis(redisCtx.Client)

	AnalyticRepo := analyticrepo.NewCassandra(
		cassandra.Session,
		cfg.Infrastructure.Database.Keyspace,
		redisCache)

	analyticsCfg := analytics.Config{
		AnalyticRepo: AnalyticRepo,
	}
	service := analytics.New(analyticsCfg)

	mux := web.NewMux(logger)

	mux.Use(web.RequestId)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))

	handlers := newHandlers(service)
	mux.Post("/v1/clicks/{alias}", handlers.Clicks)

	app := web.New(serviceCfg, logger)
	app.Run(mux)

	return 0
}
