package analyticclicks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/telemetry"
	"github.com/ClickHouse/clickhouse-go/v2"
)

type clickHouseRepo struct {
	database string
	conn     clickhouse.Conn
	cache    ports.Cache
	ttl      time.Duration
}

func NewClickHouse(database string, conn clickhouse.Conn, cache ports.Cache, ttl time.Duration) ports.AnalyticClicksRepository {
	return &clickHouseRepo{database: database, conn: conn, cache: cache, ttl: ttl}
}

func (repo *clickHouseRepo) Get(ctx context.Context, alias string) (*domain.AnalyticClicks, error) {
	dctx, span := telemetry.NewSpan(ctx, "ClickHouse Analytic Clicks Repo: Get")
	defer span.End()

	stat := domain.AnalyticClicks{Alias: alias}
	cacheErr := repo.cache.Get(dctx, alias, &stat)
	if cacheErr == nil {
		return &stat, nil
	}
	stmt := fmt.Sprintf(
		`SELECT sum(total_clicks) As clicks 
		 FROM %s.analytic_clicks_view_target FINAL
		 WHERE alias = ?
		 GROUP BY (alias)`, repo.database)
	row := repo.conn.QueryRow(dctx, stmt, alias)
	if err := row.Scan(&stat.Clicks); err != nil {
		span.RecordError(err)

		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.NewNotFound(err, "analytic clicks not found")
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errs.NewTimeout(err)
		}

		if exception, ok := err.(*clickhouse.Exception); ok {
			switch exception.Code {
			case 159:
				return nil, errs.NewTimeout(err)
			}
		}

		return nil, fmt.Errorf("'row.Scan' failed: %w", err)
	}

	if errs.IsNotFound(cacheErr) {
		if err := repo.cache.Put(ctx, alias, stat, repo.ttl); err != nil {
			logger := logging.Get(ctx)
			logger.Error("failed to set cache entry", "key", alias, "error", err)
		}
	}
	return &stat, nil
}

func (repo *clickHouseRepo) Put(ctx context.Context, ids []string, aliases []string, clickCounts []int) error {
	dctx, span := telemetry.NewSpan(ctx, "ClickHouse Analytic Clicks Repo: Put")
	defer span.End()

	stmt := fmt.Sprintf(`INSERT INTO %s.analytic_clicks`, repo.database)
	batch, err := repo.conn.PrepareBatch(dctx, stmt)
	if err != nil {
		return fmt.Errorf("'conn.PrepareBatch' failed: %w", err)
	}
	now := time.Now().UTC()
	for idx := range ids {
		id := ids[idx]
		alias := aliases[idx]
		clicks := clickCounts[idx]
		if err := batch.Append(id, alias, clicks, now); err != nil {
			return fmt.Errorf("'batch.Append' failed: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		span.RecordError(err)

		if errors.Is(err, context.DeadlineExceeded) {
			return errs.NewTimeout(err)
		}
		if exception, ok := err.(*clickhouse.Exception); ok {
			switch exception.Code {
			case 159:
				return errs.NewTimeout(err)
			}
		}
		return fmt.Errorf("'batch.Send' failed: %w", err)
	}
	return nil
}

func (repo *clickHouseRepo) Delete(ctx context.Context, alias string) error {
	dctx, span := telemetry.NewSpan(ctx, "ClickHouse Analytic Clicks Repo: Delete")
	defer span.End()

	stmt := fmt.Sprintf(
		`DELETE FROM
		%s.analytic_clicks
		WHERE alias = ?`, repo.database)
	if err := repo.conn.Exec(dctx, stmt, alias); err != nil {
		span.RecordError(err)

		if errors.Is(err, sql.ErrNoRows) {
			return errs.NewNotFound(err, "analytic clicks not found")
		}
		return fmt.Errorf("'conn.Exec' failed: %w", err)
	}
	return nil
}
