package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/async"
	analyticsV1 "github.com/AdventurerAmer/shortner/cmd/services/analytics/v1"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/caches"
	"github.com/AdventurerAmer/shortner/internal/core/services/redirecting"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmappingrepo"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	serviceCfg := &cfg.Services.Redirecting
	logger := logging.New(cfg).With(slog.String("service", serviceCfg.Name))

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

	URLMappingRepo := urlmappingrepo.NewCassandra(
		cassandra.Session,
		cfg.Infrastructure.Database.Keyspace,
		redisCache)

	redirectingCfg := redirecting.Config{
		URLMappingRepo: URLMappingRepo,
	}
	service := redirecting.New(redirectingCfg)

	analyticsClient := analyticsV1.NewClient(cfg.Services.Analytics.Address())

	orch := async.NewOrchestrator(context.Background())
	defer orch.Shutdown()

	handlers := newHandlers(service, analyticsClient, orch)

	mux := web.NewMux(logger)

	mux.Use(web.RequestId)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	mux.Get("/v1/redirect/{alias}", handlers.redirect)

	app := web.New(serviceCfg, logger)
	app.Run(mux)

	return 0
}
