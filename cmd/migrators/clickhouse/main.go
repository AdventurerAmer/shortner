package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/ClickHouse/clickhouse-go/v2"
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

	mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	path := "internal/migrations/clickhouse" // TODO: hardcoding path
	if err := runMigrations(mctx, path, clickHouse.Conn); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(ctx context.Context, path string, conn clickhouse.Conn) error {
	files, err := filepath.Glob(filepath.Join(path, "*.sql"))
	if err != nil {
		return fmt.Errorf("'filepath.Glob' failed: %w", err)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("'os.ReadFile' failed %s: %w", file, err)
		}
		if err := conn.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("'conn.Exec' failed %s: %w", file, err)
		}
	}
	return nil
}
