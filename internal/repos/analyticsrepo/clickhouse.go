package analyticsrepo

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

func NewClickHouse(database string, conn clickhouse.Conn, cache ports.Cache, ttl time.Duration) ports.AnalyticsRepository {
	return &clickHouseRepo{database: database, conn: conn, cache: cache, ttl: ttl}
}

func (repo *clickHouseRepo) Get(ctx context.Context, alias string) (*domain.Analytic, error) {
	dctx, span := telemetry.NewSpan(ctx, "ClickHouse Analytics Repo: Get")
	defer span.End()

	analytic := domain.Analytic{Alias: alias}
	cacheErr := repo.cache.Get(dctx, alias, &analytic)
	if cacheErr == nil {
		return &analytic, nil
	}
	stmt := fmt.Sprintf(
		`SELECT SUM(clicks) As total_clicks 
		 FROM %s.analytics_view_target
		 WHERE alias = ?
		 GROUP BY (alias)`, repo.database)
	row := repo.conn.QueryRow(dctx, stmt, alias)
	if err := row.Scan(&analytic.Clicks); err != nil {
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
		if err := repo.cache.Put(ctx, alias, analytic, repo.ttl); err != nil {
			logger := logging.Get(ctx)
			logger.Error("failed to set cache entry", "key", alias, "error", err)
		}
	}
	return &analytic, nil
}

func (repo *clickHouseRepo) Put(ctx context.Context, eventId, alias string, createdAt time.Time) error {
	dctx, span := telemetry.NewSpan(ctx, "ClickHouse Analytics Repo: Put")
	defer span.End()

	stmt := fmt.Sprintf(`INSERT INTO %s.analytics VALUES (?, ?, ?)`, repo.database)
	if err := repo.conn.Exec(dctx, stmt, eventId, alias, createdAt); err != nil {
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
		return fmt.Errorf("'conn.Exec' failed: %w", err)
	}
	return nil
}

func (repo *clickHouseRepo) Delete(ctx context.Context, eventId string) error {
	dctx, span := telemetry.NewSpan(ctx, "ClickHouse Analytic Clicks Repo: Delete")
	defer span.End()

	stmt := fmt.Sprintf(
		`DELETE FROM
		%s.analytics
		WHERE event_id = ?`, repo.database)
	if err := repo.conn.Exec(dctx, stmt, eventId); err != nil {
		span.RecordError(err)

		if errors.Is(err, sql.ErrNoRows) {
			return errs.NewNotFound(err, "analytic not found")
		}
		return fmt.Errorf("'conn.Exec' failed: %w", err)
	}
	return nil
}
