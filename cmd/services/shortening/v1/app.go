package v1

import (
	"fmt"
	"os"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

func Run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		return 1
	}

	logger := logging.New(cfg)

	cluster := gocql.NewCluster(cfg.Infrastructure.Database.Host)
	cluster.Keyspace = cfg.Infrastructure.Database.Keyspace
	cluster.Port = cfg.Infrastructure.Database.Port
	cluster.Consistency = gocql.Quorum
	cluster.ConnectTimeout = cfg.Infrastructure.Database.ConnTimeout

	session, err := cluster.CreateSession()
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return 1
	}
	defer session.Close()

	app := web.New(logger, &cfg.Services.Shortening)

	mux := web.NewMux()
	mux.Get("/health", app.DefaultHealthHandler)

	app.Run(mux)

	return 0
}
