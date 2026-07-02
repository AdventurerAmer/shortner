package analyticrepo

import (
	"context"
	"errors"
	"fmt"

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

func NewCassandra(session *gocql.Session, keyspace string, cache ports.Cache) ports.AnalyticRepository {
	return &cassandraRepo{session: session, keyspace: keyspace, cache: cache}
}

func (repo *cassandraRepo) Create(ctx context.Context, a *domain.Analytic) error {
	stmt := fmt.Sprintf(
		`INSERT INTO 
		 %s.analytics (alias, created_at, updated_at, clicks, version)
		 VALUES (?, ?, ?, ?, ?)`, repo.keyspace)

	q := repo.session.Query(stmt, a.Alias, a.CreatedAt, a.UpdatedAt, a.Clicks, a.Version)
	if err := q.ExecContext(ctx); err != nil {
		var (
			writeTimeout  gocql.RequestErrWriteTimeout
			alreadyExists gocql.RequestErrAlreadyExists
		)
		switch {
		case errors.As(err, &writeTimeout):
			return errs.NewTimeout(err)
		case errors.As(err, &alreadyExists):
			return errs.NewAlreadyExists(err, "analytic already exists")
		}
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}

func (repo *cassandraRepo) Get(ctx context.Context, alias string) (*domain.Analytic, error) {
	a := domain.Analytic{Alias: alias}
	// cacheErr := repo.cache.Get(ctx, alias, &a)
	// if cacheErr == nil {
	// 	return &a, nil
	// }
	stmt := fmt.Sprintf(
		`SELECT created_at, updated_at, clicks, version 
		 FROM %s.analytics
		 WHERE alias = ?`, repo.keyspace)
	query := repo.session.Query(stmt, alias).Consistency(gocql.One)
	if err := query.ScanContext(ctx, &a.CreatedAt, &a.UpdatedAt, &a.Clicks, &a.Version); err != nil {
		var timeout gocql.RequestErrReadTimeout
		switch {
		case errors.As(err, &timeout):
			return nil, errs.NewTimeout(err)
		case errors.Is(err, gocql.ErrNotFound):
			return nil, errs.NewNotFound(err, "url mapping is not found")
		}
		return nil, fmt.Errorf("'ScanContext' failed: %w", err)
	}
	// if errs.IsNotFound(cacheErr) {
	// 	ttl := 10 * time.Minute // TODO: hardcoding TTL
	// 	repo.cache.Put(ctx, alias, a, ttl)
	// }
	return &a, nil
}

func (repo *cassandraRepo) Update(ctx context.Context, a *domain.Analytic) error {
	stmt := fmt.Sprintf(
		`UPDATE %s.analytics
		 SET updated_at = ?, clicks = ?, version = ?
		 WHERE alias = ?
		 IF version = ?`, repo.keyspace)

	query := repo.session.Query(stmt, a.UpdatedAt, a.Clicks, a.Version+1, a.Alias, a.Version).Consistency(gocql.One)
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

	a.Version += 1
	return nil
}

func (repo *cassandraRepo) Delete(ctx context.Context, alias string) error {
	stmt := fmt.Sprintf(
		`DELETE FROM 
		%s.analytics 
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
