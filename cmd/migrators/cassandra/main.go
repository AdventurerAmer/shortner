package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cassandra, err := infra.ConnectToCassandra(ctx, &cfg.Infrastructure.Cassandra)
	if err != nil {
		logger.Error("cassandra connection failed", "error", err)
		os.Exit(1)
	}
	defer infra.CloseCassandra(context.TODO(), cassandra)

	logger.Info("Connected to Cassandra")

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dctx := logging.Set(sigCtx, logger)
	if err := cassandraMigratorV1.Run(dctx, &cassandra); err != nil {
		logger.Error("'cassandraMigratorV1.Run' failed", "error", err)
		os.Exit(1)
	}
}
