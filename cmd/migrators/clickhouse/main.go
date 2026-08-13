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

	logger := logging.New(cfg).With(slog.String("migrator", "clickHouse"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clickHouse, err := infra.ConnectClickHouse(ctx, &cfg.Infrastructure.ClickHouse)
	if err != nil {
		logger.Error("clickhouse connection failed", "error", err)
		os.Exit(1)
	}
	defer infra.CloseClickHouse(context.TODO(), clickHouse)

	logger.Info("Connected to Clickhouse")

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	glob := "internal/migrations/clickhouse/*.sql"
	app, err := migrator.New(logger, &clickHouse)
	if err != nil {
		logger.Error("'migrator.New' failed", "error", err)
		os.Exit(1)
	}
	if err := app.Run(sigCtx, glob); err != nil {
		logger.Error("'app.Run' failed", "error", err)
		os.Exit(1)
	}
}
