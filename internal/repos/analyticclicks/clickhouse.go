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
}

func NewClickHouse(database string, conn clickhouse.Conn, cache ports.Cache) ports.AnalyticClicksRepository {
	return &clickHouseRepo{database: database, conn: conn, cache: cache}
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
	// query := repo.session.Query(stmt, alias).Consistency(gocql.One)
	// if err := query.ScanContext(ctx, &stat.Clicks); err != nil {
	// 	var timeout gocql.RequestErrReadTimeout
	// 	switch {
	// 	case errors.As(err, &timeout):
	// 		return nil, errs.NewTimeout(err)
	// 	case errors.Is(err, gocql.ErrNotFound):
	// 		return nil, errs.NewNotFound(err, "url mapping is not found")
	// 	}
	// 	return nil, fmt.Errorf("'query.ScanContext' failed: %w", err)
	// }
	if errs.IsNotFound(cacheErr) {
		ttl := 1 * time.Second // TODO: hardcoding TTL
		repo.cache.Put(ctx, alias, stat, ttl)
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
	// query := repo.session.Query(stmt, alias)
	// if err := query.ExecContext(ctx); err != nil {
	// 	var (
	// 		readTimeout  gocql.RequestErrReadTimeout
	// 		writeTimeout gocql.RequestErrWriteTimeout
	// 	)
	// 	switch {
	// 	case errors.As(err, &readTimeout), errors.As(err, &writeTimeout):
	// 		return errs.NewTimeout(err)
	// 	case errors.Is(err, gocql.ErrNotFound):
	// 		return errs.NewNotFound(err, "analytic is not found")
	// 	}
	// 	return fmt.Errorf("'query.ExecContext' failed: %w", err)
	// }
	return nil
}
