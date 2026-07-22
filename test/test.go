package test

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/logging"
)

type Cassandra struct {
	Logger    *logging.Logger
	Cassandra infra.Cassandra
	Keyspace  string
}

func NewCassandraTestContext() (*Cassandra, error) {
	if err := ChangeToRootDir(); err != nil {
		return nil, fmt.Errorf("'ChangeToRootDir' failed: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("'config.Load' failed: %w", err)
	}

	logger := logging.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cassandra, err := infra.ConnectToCassandra(ctx, &cfg.Infrastructure.Cassandra)
	if err != nil {
		return nil, fmt.Errorf("'infra.ConnectToCassandra' failed: %w", err)
	}

	return &Cassandra{
		Logger:    logger,
		Cassandra: cassandra,
		Keyspace:  cfg.Infrastructure.Cassandra.Keyspace,
	}, nil
}

func (c *Cassandra) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	infra.CloseCassandra(ctx, c.Cassandra)
}

type ClickHouse struct {
	Logger     *logging.Logger
	ClickHouse infra.ClickHouse
	Database   string
}

func NewClickHouseTestContext() (*ClickHouse, error) {
	if err := ChangeToRootDir(); err != nil {
		return nil, fmt.Errorf("'ChangeToRootDir' failed: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("'config.Load' failed: %w", err)
	}

	logger := logging.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clickHouse, err := infra.ConnectClickHouse(ctx, &cfg.Infrastructure.ClickHouse)
	if err != nil {
		return nil, fmt.Errorf("'infra.ConnectClickHouse' failed: %w", err)
	}

	return &ClickHouse{
		Logger:     logger,
		ClickHouse: clickHouse,
		Database:   cfg.Infrastructure.ClickHouse.Database,
	}, nil
}

func (c *ClickHouse) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	infra.CloseClickHouse(ctx, c.ClickHouse)
}
