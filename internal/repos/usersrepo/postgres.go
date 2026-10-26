package usersrepo

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

func NewPostgres(db *sql.DB) ports.UsersRepository {
	return &postgresRepo{
		db: db,
	}
}

func (repo *postgresRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, password_hash, display_name, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	row := postgresutils.QueryRow(ctx, repo.db, query, user.Email, user.PasswordHash, user.DisplayName, user.IsVerified, user.CreatedAt, user.CreatedAt)
	if err := row.Scan(&user.Id); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}

	return nil
}

func (repo *postgresRepo) Get(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT email, password_hash, display_name, is_verified, created_at, updated_at, version 
		FROM users 
		WHERE id = $1`

	row := postgresutils.QueryRow(ctx, repo.db, query, id)

	var user domain.User
	if err := row.Scan(&user.Email, &user.PasswordHash, &user.DisplayName, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt, &user.Version); err != nil {
		return nil, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return &user, nil
}

func (repo *postgresRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, password_hash, display_name, is_verified, created_at, updated_at, version 
		FROM users 
		WHERE email = $1`
	row := postgresutils.QueryRow(ctx, repo.db, query, email)
	var user domain.User
	if err := row.Scan(&user.Id, &user.PasswordHash, &user.DisplayName, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt, &user.Version); err != nil {
		return nil, fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return &user, nil
}

func (repo *postgresRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, display_name = $3, is_verified = $4, updated_at = $5, version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version
	`
	row := postgresutils.QueryRow(ctx, repo.db, query, user.Email, user.PasswordHash, user.DisplayName, user.IsVerified, user.UpdatedAt, user.Id, user.Version)
	if err := row.Scan(&user.Version); err != nil {
		return fmt.Errorf("'row.Scan' failed: %w", err)
	}
	return nil
}

func (repo *postgresRepo) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`
	if _, err := postgresutils.Exec(ctx, repo.db, query, id); err != nil {
		return fmt.Errorf("'postgresutils.Exec' failed: %w", err)
	}
	return nil
}
