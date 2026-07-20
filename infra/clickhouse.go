package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouse struct {
	Conn clickhouse.Conn
}

func ConnectClickHouse(ctx context.Context, cfg *config.ClickHouseConfig) (ClickHouse, error) {
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
		return ClickHouse{}, ctx.Err()
	case res := <-ch:
		return res.clickHouse, res.err
	}
}

func CloseClickHouse(ctx context.Context, clickHouse ClickHouse) error {
	errCh := make(chan error)
	go func() {
		if err := clickHouse.Conn.Close(); err != nil {
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
