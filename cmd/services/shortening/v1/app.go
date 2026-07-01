package v1

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/core/services/shortening"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmappingrepo"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/web"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	serviceCfg := &cfg.Services.Shortening
	logger := logging.New(cfg).With(slog.String("service", serviceCfg.Name))

	cassandra, err := infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
	if err != nil {
		logger.Error("cassandra connection failed", "error", err)
		return 1
	}
	defer infra.CloseCassandra(context.TODO(), cassandra)

	urlmappingRepo := urlmappingrepo.NewCassandra(cassandra.Session, cfg.Infrastructure.Database.Keyspace, ports.NewCacheStub())

	idGenerator := snowflake.New("sa")
	shorteningCfg := shortening.Config{
		ShortURLPrefix: "http://localhost:3031/v1/redirect/",
		URLMappingRepo: urlmappingRepo,
		IdGenerator:    idGenerator,
	}
	service := shortening.New(shorteningCfg)

	handlers := NewHandlers(service)

	mux := web.NewMux(logger)
	mux.Use(web.RequestId)
	mux.Use(web.Logging)
	mux.Use(web.Recover(cfg.Env))
	mux.Use(web.Timeout(serviceCfg.DefaultTimeout))

	mux.Post("/v1/shorten", handlers.shorten)

	app := web.New(serviceCfg, logger)
	app.Run(mux)

	return 0
}
