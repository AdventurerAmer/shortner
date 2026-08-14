package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	clickhouseMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/clickhouse/v1"
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

	dctx := logging.Set(sigCtx, logger)
	if err := clickhouseMigratorV1.Run(dctx, &clickHouse); err != nil {
		logger.Error("'clickhouseMigratorV1.Run' failed", "error", err)
		os.Exit(1)
	}
}
