package urlmappingrepo

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type cassandraRepo struct {
	session *gocql.Session
}

func NewCassandra(session *gocql.Session) ports.URLMappingRepository {
	return &cassandraRepo{session: session}
}

func (repo *cassandraRepo) Create(ctx context.Context, m *domain.URLMapping) error {
	stmt := `INSERT INTO 
			 url_mappings (short_url, long_url, created_at, user_id)
			 VALUES (?, ?, ?, ?)`

	q := repo.session.Query(stmt, m.ShortURL, m.LongURL, m.CreatedAt, m.UserId)
	if err := q.ExecContext(ctx); err != nil {
		return err
	}

	return nil
}

func (repo *cassandraRepo) Get(ctx context.Context, shortURL string) (*domain.URLMapping, error) {
	stmt := `SELECT long_url, created_at, user_id 
			 FROM url_mappings
			 WHERE short_url = ?`
	mapping := domain.URLMapping{ShortURL: shortURL}
	query := repo.session.Query(stmt, shortURL).Consistency(gocql.One)
	if err := query.ScanContext(ctx, &mapping.ShortURL, &mapping.CreatedAt, &mapping.UserId); err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (repo *cassandraRepo) Delete(ctx context.Context, shortURL string) error {
	stmt := `DELETE FROM 
			 url_mappings 
			 WHERE short_url = ?`
	query := repo.session.Query(stmt, shortURL)
	if err := query.ExecContext(ctx); err != nil {
		return err
	}
	return nil
}
