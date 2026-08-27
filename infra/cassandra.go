package infra

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type cassandraConfigWrapper struct {
	config.CassandraConfig
}

func (cfg *cassandraConfigWrapper) Connect(ctx context.Context) (Disconnecter, error) {
	type result struct {
		cassandra Cassandra
		err       error
	}
	ch := make(chan result)
	go func() {
		cluster := gocql.NewCluster(cfg.Host)
		cluster.Port = cfg.Port
		cluster.Consistency = gocql.Quorum
		cluster.ProtoVersion = 4
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		session, err := cluster.CreateSession()
		if err != nil {
			err = fmt.Errorf("'cluster.CreateSession' failed: %w", err)
		}
		c := Cassandra{
			Session: session,
		}
		ch <- result{cassandra: c, err: err}
	}()

	select {
	case res := <-ch:
		return &res.cassandra, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type Cassandra struct {
	Session *gocql.Session
}

func (c *Cassandra) Disconnect(ctx context.Context) error {
	c.Session.Close()
	return nil
}

func (c *Cassandra) Execute(ctx context.Context, query string) error {
	if err := c.Session.Query(query).ExecContext(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errs.NewTimeout(err)
		}
		return fmt.Errorf("'session.Query' failed: %w", err)
	}
	return nil
}

func (c *Cassandra) Ping(ctx context.Context) error {
	var releaseVersion string
	// system.local is always accessible by any authenticated user
	query := c.Session.Query("SELECT release_version FROM system.local").
		Consistency(gocql.One)
	if err := query.ScanContext(ctx, &releaseVersion); err != nil {
		return fmt.Errorf("'query.ScanContext' failed: %w", err)
	}
	return nil
}
