package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/ClickHouse/clickhouse-go/v2"
)

type clickHouseConfigWrapper struct {
	config.ClickHouseConfig
}

func (cfg *clickHouseConfigWrapper) Connect(ctx context.Context) (Disconnecter, error) {
	type result struct {
		clickHouse ClickHouse
		err        error
	}
	ch := make(chan result)
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		opts := &clickhouse.Options{
			Addr: []string{addr},
			Auth: clickhouse.Auth{
				Database: cfg.Database,
				Username: cfg.Username,
				Password: cfg.Password,
			},
		}
		conn, err := clickhouse.Open(opts)
		res := result{
			clickHouse: ClickHouse{
				Conn: conn,
			},
			err: err,
		}
		ch <- res
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return &res.clickHouse, res.err
	}
}

type ClickHouse struct {
	Conn clickhouse.Conn
}

func (ch *ClickHouse) Disconnect(ctx context.Context) error {
	errCh := make(chan error)
	go func() {
		if err := ch.Conn.Close(); err != nil {
			errCh <- fmt.Errorf("'Conn.Close' failed: %w", err)
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (ch *ClickHouse) Execute(ctx context.Context, query string) error {
	if err := ch.Conn.Exec(ctx, query); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errs.NewTimeout(err)
		}
		return fmt.Errorf("'Conn.Exec' failed %s: %w", query, err)
	}
	return nil
}
