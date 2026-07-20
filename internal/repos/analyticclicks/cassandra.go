package analyticclicks

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

func NewCassandra(session *gocql.Session, keyspace string, cache ports.Cache) ports.AnalyticClicksRepository {
	return &cassandraRepo{session: session, keyspace: keyspace, cache: cache}
}

func (repo *cassandraRepo) Get(ctx context.Context, alias string) (*domain.AnalyticClicks, error) {
	stat := domain.AnalyticClicks{Alias: alias}
	cacheErr := repo.cache.Get(ctx, alias, &stat)
	if cacheErr == nil {
		return &stat, nil
	}
	stmt := fmt.Sprintf(
		`SELECT clicks 
		 FROM %s.analytic_clicks
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

func (repo *cassandraRepo) Put(ctx context.Context, ids []string, aliases []string, clicks []int) error {
	batch := repo.session.Batch(gocql.LoggedBatch)

	insertStmt := fmt.Sprintf(`
		INSERT INTO 
		%s.analytic_click_batches (id, alias, clicks, applied_at)
		VALUES (?, ?, ?, ?)
	`, repo.keyspace)
	now := time.Now().UTC()
	for i := range len(aliases) {
		batch.Query(insertStmt, ids[i], aliases[i], clicks[i], now)
	}

	updateStmt := fmt.Sprintf(
		`UPDATE %s.analytic_clicks
		 SET clicks = clicks + ?
		 WHERE alias = ?`, repo.keyspace)
	for i := range len(aliases) {
		batch.Query(updateStmt, clicks[i], aliases[i])
	}

	if err := batch.ExecContext(ctx); err != nil {
		var (
			readTimeout   gocql.RequestErrReadTimeout
			writeTimeout  gocql.RequestErrWriteTimeout
			alreadyExists gocql.RequestErrAlreadyExists
		)
		switch {
		case errors.As(err, &readTimeout), errors.As(err, &writeTimeout):
			return errs.NewTimeout(err)
		case errors.As(err, &alreadyExists):
			return errs.NewAlreadyExists(err, "clicks batch already exists")
		}
		return fmt.Errorf("'batch.ExecContext' failed: %w", err)
	}
	return nil
}

func (repo *cassandraRepo) Delete(ctx context.Context, alias string) error {
	stmt := fmt.Sprintf(
		`DELETE FROM 
		%s.analytic_clicks 
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
		return fmt.Errorf("'query.ExecContext' failed: %w", err)
	}
	return nil
}
