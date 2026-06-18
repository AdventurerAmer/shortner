package urlmappingrepo

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
}

func NewCassandra(session *gocql.Session, keyspace string) ports.URLMappingRepository {
	return &cassandraRepo{session: session, keyspace: keyspace}
}

func (repo *cassandraRepo) Create(ctx context.Context, m *domain.URLMapping) error {
	stmt := fmt.Sprintf(
		`INSERT INTO 
		 %s.url_mappings (alias, long_url, created_at, user_id)
		 VALUES (?, ?, ?, ?)`, repo.keyspace)

	q := repo.session.Query(stmt, m.Alias, m.LongURL, m.CreatedAt, m.UserId)
	if err := q.ExecContext(ctx); err != nil {
		var (
			writeTimeout  gocql.RequestErrWriteTimeout
			alreadyExists gocql.RequestErrAlreadyExists
		)
		switch {
		case errors.As(err, &writeTimeout):
			return errs.NewTimeout(err)
		case errors.As(err, &alreadyExists):
			return errs.NewAlreadyExists(err, "url mapping already exists")
		}
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}

func (repo *cassandraRepo) Get(ctx context.Context, alias string) (*domain.URLMapping, error) {
	stmt := fmt.Sprintf(
		`SELECT long_url, created_at, user_id 
		 FROM %s.url_mappings
		 WHERE alias = ?`, repo.keyspace)
	m := domain.URLMapping{Alias: alias}
	query := repo.session.Query(stmt, alias).Consistency(gocql.One)
	if err := query.ScanContext(ctx, &m.LongURL, &m.CreatedAt, &m.UserId); err != nil {
		var timeout gocql.RequestErrReadTimeout
		switch {
		case errors.As(err, &timeout):
			return nil, errs.NewTimeout(err)
		case errors.Is(err, gocql.ErrNotFound):
			return nil, errs.NewNotFound(err, "url mapping is not found")
		}
		return nil, fmt.Errorf("'ScanContext' failed: %w", err)
	}
	return &m, nil
}

func (repo *cassandraRepo) Delete(ctx context.Context, alias string) error {
	stmt := fmt.Sprintf(
		`DELETE FROM 
		%s.url_mappings 
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
			return errs.NewNotFound(err, "url mapping is not found")
		}
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}
	return nil
}
