package analyticclicks

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
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
	stat := domain.AnalyticClicks{Alias: alias}
	cacheErr := repo.cache.Get(ctx, alias, &stat)
	if cacheErr == nil {
		return &stat, nil
	}
	stmt := fmt.Sprintf(
		`SELECT sum(total_clicks) AS clicks
		 FROM %s.analytic_clicks_view
		 WHERE alias = ?
		 GROUP BY alias`, repo.database)
	row := repo.conn.QueryRow(ctx, stmt, alias)
	if err := row.Scan(&stat.Clicks); err != nil {
		return nil, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	if errs.IsNotFound(cacheErr) {
		repo.cache.Put(ctx, alias, stat, repo.ttl)
	}
	return &stat, nil
}

func (repo *clickHouseRepo) Put(ctx context.Context, ids []string, aliases []string, clickCounts []int) error {
	stmt := fmt.Sprintf(`INSERT INTO %s.analytic_clicks`, repo.database)
	batch, err := repo.conn.PrepareBatch(ctx, stmt)
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
		return fmt.Errorf("'batch.Send' failed: %w", err)
	}
	return nil
}

func (repo *clickHouseRepo) Delete(ctx context.Context, alias string) error {
	stmt := fmt.Sprintf(
		`DELETE FROM
		%s.analytic_clicks
		WHERE alias = ?`, repo.database)
	if err := repo.conn.Exec(ctx, stmt, alias); err != nil {
		return fmt.Errorf("'conn.Exec' failed: %w", err)
	}
	return nil
}
