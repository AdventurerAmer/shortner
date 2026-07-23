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
	gocql "github.com/apache/cassandra-gocql-driver/v2"
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

	path := "internal/migrations/cassandra" // TODO: hardcoding path
	if err := runMigrations(context.TODO(), cassandra.Session, path, logger); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(ctx context.Context, session *gocql.Session, path string, logger *logging.Logger) error {
	files, err := filepath.Glob(filepath.Join(path, "*.cql"))
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
			logger.Info("Executing Query", "file", file, "query", q)
			if err := execQuery(ctx, session, q); err != nil {
				return fmt.Errorf("'execQuery' failed %q:%q: %w", file, q, err)
			}
		}
	}
	return nil
}

func execQuery(ctx context.Context, session *gocql.Session, query string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := session.Query(query).ExecContext(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errs.NewTimeout(err)
		}
		return fmt.Errorf("'session.Query' failed: %w", err)
	}

	return nil
}
