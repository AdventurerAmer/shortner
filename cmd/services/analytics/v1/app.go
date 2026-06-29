package v1

import (
	"context"
	"fmt"
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

	logger := logging.New(cfg)

	cassandra, err := infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
	if err != nil {
		logger.Error("cassandra connection failed", "error", err)
		return 1
	}
	defer infra.CloseCassandra(context.TODO(), cassandra)

	redisCtx, err := infra.ConnectToRedis(context.TODO(), &cfg.Infrastructure.Redis)
	if err != nil {
		logger.Error("redis connection failed", "error", err)
		return 1
	}
	defer infra.CloseRedis(context.TODO(), redisCtx)

	redisCache := caches.NewRedis(redisCtx.Client)

	app := web.New(cfg.Env, logger, &cfg.Services.Analytics)

	AnalyticRepo := analyticrepo.NewCassandra(
		cassandra.Session,
		cfg.Infrastructure.Database.Keyspace,
		redisCache)

	analyticsCfg := analytics.Config{
		AnalyticRepo: AnalyticRepo,
	}
	srv := analytics.New(analyticsCfg)

	handlers := NewHandlers(logger, &cfg.Services.Analytics, srv)

	mux := web.NewMux(logger)

	mux.Use(app.RequestId)
	mux.Use(app.Logging)
	mux.Use(app.Recover)

	mux.Get("/health", app.DefaultHealthHandler)
	mux.Get("/v1/clicks/{alias}", handlers.Clicks)

	app.Run(mux)

	return 0
}
