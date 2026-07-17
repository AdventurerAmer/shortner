package analyticstat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type cassandraRepo struct {
	session  *gocql.Session
	keyspace string
	cache    ports.Cache
}

func NewCassandra(session *gocql.Session, keyspace string, cache ports.Cache) ports.AnalyticStatRepository {
	return &cassandraRepo{session: session, keyspace: keyspace, cache: cache}
}

func (repo *cassandraRepo) Get(ctx context.Context, alias string) (*domain.AnalyticStat, error) {
	stat := domain.AnalyticStat{Alias: alias}
	cacheErr := repo.cache.Get(ctx, alias, &stat)
	if cacheErr == nil {
		return &stat, nil
	}
	stmt := fmt.Sprintf(
		`SELECT clicks 
		 FROM %s.analytic_stats
		 WHERE alias = ?`, repo.keyspace)
	query := repo.session.Query(stmt, alias).Consistency(gocql.One)
	if err := query.ScanContext(ctx, &stat.Clicks); err != nil {
		var timeout gocql.RequestErrReadTimeout
		switch {
		case errors.As(err, &timeout):
			return nil, errs.NewTimeout(err)
		case errors.Is(err, gocql.ErrNotFound):
			return nil, errs.NewNotFound(err, "url mapping is not found")
		}
		return nil, fmt.Errorf("'query.ScanContext' failed: %w", err)
	}
	if errs.IsNotFound(cacheErr) {
		ttl := 1 * time.Second // TODO: hardcoding TTL
		repo.cache.Put(ctx, alias, stat, ttl)
	}
	return &stat, nil
}

func (repo *cassandraRepo) Put(ctx context.Context, id string, aliases []string, clicks []int) error {
	// insertStmt := fmt.Sprintf(`
	// 	INSERT INTO %s.analytic_stats_batches (batch_id, applied_at)
	// 	VALUES (?, ?) IF NOT EXISTS
	// `, repo.keyspace)
	// batch.Query(stmt, id, time.Now().UTC())

	stmt := fmt.Sprintf(
		`UPDATE %s.analytic_stats
		 SET clicks = clicks + ?
		 WHERE alias = ?`, repo.keyspace)

	batch := repo.session.Batch(gocql.CounterBatch)
	for i := range len(aliases) {
		batch.Query(stmt, clicks[i], aliases[i])
	}

	if err := batch.ExecContext(ctx); err != nil {
		var (
			readTimeout  gocql.RequestErrReadTimeout
			writeTimeout gocql.RequestErrWriteTimeout
		)
		switch {
		case errors.As(err, &readTimeout), errors.As(err, &writeTimeout):
			return errs.NewTimeout(err)
		}
		return fmt.Errorf("'batch.ExecContext' failed: %w", err)
	}
	return nil
}

func (repo *cassandraRepo) Delete(ctx context.Context, alias string) error {
	stmt := fmt.Sprintf(
		`DELETE FROM 
		%s.analytic_stats 
		WHERE alias = ?`, repo.keyspace)

	query := repo.session.Query(stmt, alias)
	if err := query.ExecContext(ctx); err != nil {
		var (
			readTimeout  gocql.RequestErrReadTimeout
			writeTimeout gocql.RequestErrWriteTimeout
		)
		switch {
		case errors.As(err, &readTimeout), errors.As(err, &writeTimeout):
			return errs.NewTimeout(err)
		case errors.Is(err, gocql.ErrNotFound):
			return errs.NewNotFound(err, "analytic is not found")
		}
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}
	return nil
}
