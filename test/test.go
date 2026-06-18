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

	cassandra, err := infra.ConnectToCassandra(ctx, &cfg.Infrastructure.Database)
	if err != nil {
		return nil, fmt.Errorf("'infra.ConnectToCassandra' failed: %w", err)
	}

	return &Cassandra{
		Logger:    logger,
		Cassandra: cassandra,
		Keyspace:  cfg.Infrastructure.Database.Keyspace,
	}, nil
}

func (c *Cassandra) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	infra.CloseCassandra(ctx, c.Cassandra)
}
