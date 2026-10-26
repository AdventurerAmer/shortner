package postgresutils

import (
	"context"
	"database/sql"
)

func QueryRow(ctx context.Context, db *sql.DB, query string, args ...any) *sql.Row {
	tx := txFromContext(ctx)
	if tx == nil {
		return db.QueryRowContext(ctx, query, args)
	}
	return tx.QueryRowContext(ctx, query, args)
}

func Exec(ctx context.Context, db *sql.DB, query string, args ...any) (sql.Result, error) {
	tx := txFromContext(ctx)
	if tx == nil {
		return db.ExecContext(ctx, query, args)
	}
	return tx.ExecContext(ctx, query, args)
}
