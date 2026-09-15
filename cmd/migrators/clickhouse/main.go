package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	clickhouseMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/clickhouse/v1"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger := logging.New()
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := cfg.NewLogger().With(slog.String("migrator", "clickHouse"))

	var clickHouseCtx infra.ClickHouse

	inf, err := infra.New()
	if err != nil {
		logger.Error("'infra.New()' failed", "error", err)
		os.Exit(1)
	}
	inf.BindClickHouse(cfg.Infra.ClickHouse, &clickHouseCtx)

	if err := inf.Start(context.Background()); err != nil {
		logger.Error("infrastructure connection failed", "error", err)
		os.Exit(1)
	}
	defer inf.Shutdown(context.Background())

	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dctx := logging.Set(sigCtx, logger)
	if err := clickhouseMigratorV1.Run(dctx, &clickHouseCtx); err != nil {
		logger.Error("'clickhouseMigratorV1.Run' failed", "error", err)
		os.Exit(1)
	}
}
