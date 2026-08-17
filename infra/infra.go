package infra

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/AdventurerAmer/shortner/config"
)

type Connecter interface {
	Connect(ctx context.Context) (Disconnecter, error)
}

type Disconnecter interface {
	Disconnect(ctx context.Context) error
}

type Infra struct {
	Config
	mappings map[Connecter]Disconnecter
}

func New(opts ...Option) (*Infra, error) {
	cfg := Config{
		startupTimeout:  2 * time.Second,
		shutdownTimeout: 2 * time.Second,
	}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	infra := &Infra{
		Config:   cfg,
		mappings: make(map[Connecter]Disconnecter),
	}
	return infra, nil
}

func (infra *Infra) Bind(connector Connecter, disconnector Disconnecter) {
	if _, ok := infra.mappings[connector]; ok {
		panic("connector is already bound")
	}
	infra.mappings[connector] = disconnector
}

func (infra *Infra) Start(ctx context.Context) error {
	dctx, cancel := context.WithTimeout(ctx, infra.startupTimeout)
	defer cancel()

	errCh := make(chan error)

	wg := sync.WaitGroup{}
	done := make(chan struct{})

	for conn, dstDisconn := range infra.mappings {
		wg.Go(func() {
			srcDisconn, err := conn.Connect(dctx)
			if err != nil {
				errCh <- err
			} else {
				ele := reflect.ValueOf(dstDisconn).Elem()
				ele.Set(reflect.ValueOf(srcDisconn).Elem())
			}
		})
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		case err := <-errCh:
			return err
		}
	}
}

func (infra *Infra) Shutdown(ctx context.Context) {
	dctx, cancel := context.WithTimeout(ctx, infra.shutdownTimeout)
	defer cancel()

	wg := sync.WaitGroup{}
	done := make(chan struct{})
	errCh := make(chan error)

	for _, disconn := range infra.mappings {
		wg.Go(func() {
			if err := disconn.Disconnect(dctx); err != nil {
				errCh <- err
			}
		})
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case err := <-errCh:
			slog.Error("infrastructure shutdown failed", "error", err)
		}
	}
}

func (infra *Infra) BindCassandra(cfg config.CassandraConfig, c *Cassandra) {
	wrapper := &cassandraConfigWrapper{CassandraConfig: cfg}
	infra.Bind(wrapper, c)
}

func (infra *Infra) BindClickHouse(cfg config.ClickHouseConfig, ch *ClickHouse) {
	wrapper := &clickHouseConfigWrapper{ClickHouseConfig: cfg}
	infra.Bind(wrapper, ch)
}

func (infra *Infra) BindRedis(cfg config.RedisConfig, r *Redis) {
	wrapper := &redisConfigWrapper{RedisConfig: cfg}
	infra.Bind(wrapper, r)
}
