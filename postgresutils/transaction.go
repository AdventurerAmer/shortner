package postgresutils

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
)

type transactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) ports.Transactor {
	return &transactor{
		db: db,
	}
}

type txnContextKey struct{}

func txFromContext(ctx context.Context) *sql.Tx {
	tx, _ := ctx.Value(txnContextKey{}).(*sql.Tx)
	return tx
}

func (t *transactor) WithTransaction(ctx context.Context, fn ports.TxnFunc) error {
	tx := txFromContext(ctx)
	if tx != nil {
		return fn(ctx)
	}

	opts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted, // Default for PostgreSQL
	}
	tx, err := t.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("'db.BeginTx' failed: %w", err)
	}

	dctx := context.WithValue(ctx, txnContextKey{}, tx)
	if err := fn(dctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("'tx.Commit' failed: %w", err)
	}
	return nil
}
