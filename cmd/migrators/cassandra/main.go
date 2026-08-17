package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	cassandraMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/cassandra/v1"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg).With(slog.String("migrator", "cassandra"))

	var cassandraCtx infra.Cassandra

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		os.Exit(1)
	}
	inf.BindCassandra(cfg.Infrastructure.Cassandra, &cassandraCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		os.Exit(1)
	}
	defer inf.Shutdown(context.Background())

	logger.Info("Connected to Cassandra")

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dctx := logging.Set(sigCtx, logger)
	if err := cassandraMigratorV1.Run(dctx, &cassandraCtx); err != nil {
		logger.Error("'cassandraMigratorV1.Run' failed", "error", err)
		os.Exit(1)
	}
}
