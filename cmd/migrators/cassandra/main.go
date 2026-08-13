package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdventurerAmer/shortner/apps/migrator"
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

	glob := "internal/migrations/cassandra/*.cql"
	app, err := migrator.New(logger, &cassandra)
	if err != nil {
		logger.Error("'migrator.New' failed", "error", err)
		os.Exit(1)
	}
	if err := app.Run(sigCtx, glob); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		os.Exit(1)
	}
}
