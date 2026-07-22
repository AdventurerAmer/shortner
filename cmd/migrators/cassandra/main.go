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

	mctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	path := "internal/migrations/cassandra" // TODO: hardcoding path
	if err := runMigrations(mctx, path, cassandra.Session); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(ctx context.Context, path string, session *gocql.Session) error {
	files, err := filepath.Glob(filepath.Join(path, "*.cql"))
	if err != nil {
		return fmt.Errorf("'filepath.Glob' failed: %w", err)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("'os.ReadFile' failed %s: %w", file, err)
		}
		if err := session.Query(string(content)).ExecContext(ctx); err != nil {
			return fmt.Errorf("'session.Query' failed %s: %w", file, err)
		}
	}
	return nil
}
