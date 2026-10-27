package tokensrepo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/postgresutils"
)

type postgresRepo struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) ports.TokensRepository {
	return &postgresRepo{
		db: db,
	}
}

func (repo *postgresRepo) Create(ctx context.Context, token *domain.Token) error {
	query := `
		INSERT INTO tokens (type, user_id, hash, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	row := postgresutils.QueryRow(ctx, repo.db, query, token.Type, token.UserId, token.Hash, token.CreatedAt, token.ExpiresAt)
	if err := row.Scan(&token.Id); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}

	return nil
}

func (repo *postgresRepo) Get(ctx context.Context, tokenType domain.TokenType, hash string) (*domain.Token, error) {
	query := `
		SELECT id, user_id, created_at, expires_at, used_at 
		FROM tokens
		WHERE type = $1 AND hash = $2`

	row := postgresutils.QueryRow(ctx, repo.db, query, tokenType, hash)

	var token domain.Token
	if err := row.Scan(&token.Id, &token.UserId, &token.CreatedAt, &token.ExpiresAt, &token.UsedAt); err != nil {
		return nil, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return &token, nil
}

func (repo *postgresRepo) GetLast(ctx context.Context, userId string, tokenType domain.TokenType) (*domain.Token, error) {
	return nil, nil
}

func (repo *postgresRepo) Update(ctx context.Context, token *domain.Token) error {
	query := `
		UPDATE tokens
		SET used_at = $1
		WHERE id = $2
	`
	if _, err := postgresutils.Exec(ctx, repo.db, query, token.Id, token.UsedAt); err != nil {
		return fmt.Errorf("'postgresutils.Exec' failed: %w", err)
	}
	return nil
}

func (repo *postgresRepo) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM tokens
		WHERE id = $1
	`
	if _, err := postgresutils.Exec(ctx, repo.db, query, id); err != nil {
		return fmt.Errorf("'postgresutils.Exec' failed: %w", err)
	}
	return nil
}
