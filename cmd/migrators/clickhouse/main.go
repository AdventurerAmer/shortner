package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
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

	path := "internal/migrations/clickhouse" // TODO: hardcoding path
	if err := runMigrations(context.TODO(), clickHouse.Conn, path, logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(ctx context.Context, conn clickhouse.Conn, path string, logger *logging.Logger) error {
	files, err := filepath.Glob(filepath.Join(path, "*.sql"))
	if err != nil {
		return fmt.Errorf("'filepath.Glob' failed: %w", err)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("'os.ReadFile' failed %s: %w", file, err)
		}
		queries := strings.Split(string(content), ";")
		for _, query := range queries {
			q := strings.TrimSpace(query)
			if q == "" || strings.HasPrefix(query, "--") || strings.HasPrefix(query, "//") {
				continue
			}
			logger.Info("Executing Query", "file", file)
			if err := execQuery(ctx, conn, q); err != nil {
				return fmt.Errorf("'execQuery' failed %q: %w", q, err)
			}
		}
	}
	return nil
}

func execQuery(ctx context.Context, conn clickhouse.Conn, query string) error {
	dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := conn.Exec(dctx, query); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errs.NewTimeout(err)
		}
		return fmt.Errorf("'conn.Exec' failed %s: %w", query, err)
	}

	return nil
}
