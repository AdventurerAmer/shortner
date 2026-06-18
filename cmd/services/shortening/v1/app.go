package v1

import (
	"context"
	"fmt"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/services/shortening"
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

	logger := logging.New(cfg)

	cassandra, err := infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
	defer infra.CloseCassandra(context.TODO(), cassandra)

	app := web.New(cfg.Env, logger, &cfg.Services.Shortening)

	urlmappingRepo := urlmappingrepo.NewCassandra(cassandra.Session, cfg.Infrastructure.Database.Keyspace)

	snowflake := domain.NewSnowflake()

	shorteningCfg := shortening.Config{
		ShortURLPrefix: "http://localhost:3031/v1/redirect/",
		Shard:          "sa", // TODO: hardcoding shard
		URLMappingRepo: urlmappingRepo,
		Snowflake:      snowflake,
	}
	shorteningSrv := shortening.New(shorteningCfg)

	handlers := NewHandlers(&cfg.Services.Shortening, shorteningSrv)

	mux := web.NewMux(logger)

	mux.Use(app.RequestId)
	mux.Use(app.Logging)
	mux.Use(app.Recover)

	mux.Get("/health", app.DefaultHealthHandler)
	mux.Post("/v1/shorten", handlers.Shorten)

	app.Run(mux)

	return 0
}
