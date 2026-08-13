package migrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
)

type App struct {
	Config
	logger   *logging.Logger
	migrator ports.Migrator
}

func New(logger *logging.Logger, migrator ports.Migrator, opts ...Option) (*App, error) {
	cfg := Config{
		queryTimeout: 2 * time.Second,
	}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	app := &App{
		Config:   cfg,
		logger:   logger,
		migrator: migrator,
	}
	return app, nil
}

func (app *App) Run(ctx context.Context, glob string) error {
	files, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("'filepath.Glob' failed: %w", err)
	}
	for _, file := range files {
		app.logger.Info("Reading Contents", "file", file)
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
			app.logger.Info("Running Statement", "query", q)
			if err := app.runQuery(ctx, q); err != nil {
				return fmt.Errorf("'runQuery' failed: %w", err)
			}
		}
	}
	return nil
}

func (app *App) runQuery(ctx context.Context, query string) error {
	dctx, cancel := context.WithTimeout(ctx, app.queryTimeout)
	defer cancel()
	if err := app.migrator.Execute(dctx, query); err != nil {
		return fmt.Errorf("'migrator.Execute' failed %q: %w", query, err)
	}
	return nil
}
