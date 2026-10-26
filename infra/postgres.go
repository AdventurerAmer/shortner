package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresConfigWrapper struct {
	config.Postgres
}

func (cfg *postgresConfigWrapper) Connect(ctx context.Context) (Disconnecter, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("'sql.Open' failed: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdelConnections)
	db.SetConnMaxIdleTime(cfg.MaxIdelTime)

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &Postgres{Database: db}, nil
}

type Postgres struct {
	Database *sql.DB
}

func (p *Postgres) Disconnect(ctx context.Context) error {
	if err := p.Database.Close(); err != nil {
		return fmt.Errorf("'Database.Close' failed: %w", err)
	}
	return nil
}

func (p *Postgres) Execute(ctx context.Context, query string) error {
	if _, err := p.Database.ExecContext(ctx, query); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errs.NewTimeout(err)
		}
		return fmt.Errorf("'Database.ExecContext' failed: %w", err)
	}
	return nil
}
