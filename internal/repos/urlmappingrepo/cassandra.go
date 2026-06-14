package urlmappingrepo

import (
	"context"
	"fmt"

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
		 %s.url_mappings (short_url, long_url, created_at, user_id)
		 VALUES (?, ?, ?, ?)`, repo.keyspace)

	q := repo.session.Query(stmt, m.ShortURL, m.LongURL, m.CreatedAt, m.UserId)
	if err := q.ExecContext(ctx); err != nil {
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}

	return nil
}

func (repo *cassandraRepo) Get(ctx context.Context, shortURL string) (*domain.URLMapping, error) {
	stmt := fmt.Sprintf(
		`SELECT long_url, created_at, user_id 
		 FROM %s.url_mappings
		 WHERE short_url = ?`, repo.keyspace)
	mapping := domain.URLMapping{ShortURL: shortURL}
	query := repo.session.Query(stmt, shortURL).Consistency(gocql.One)
	if err := query.ScanContext(ctx, &mapping.LongURL, &mapping.CreatedAt, &mapping.UserId); err != nil {
		return nil, fmt.Errorf("'ScanContext' failed: %w", err)
	}
	return &mapping, nil
}

func (repo *cassandraRepo) Delete(ctx context.Context, shortURL string) error {
	stmt := fmt.Sprintf(
		`DELETE FROM 
		%s.url_mappings 
		WHERE short_url = ?`, repo.keyspace)
	query := repo.session.Query(stmt, shortURL)
	if err := query.ExecContext(ctx); err != nil {
		return fmt.Errorf("'ExecContext' failed: %w", err)
	}
	return nil
}
