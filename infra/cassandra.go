package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type Cassandra struct {
	Session *gocql.Session
}

func ConnectToCassandra(ctx context.Context, cfg *config.CassandraConfig) (Cassandra, error) {
	type result struct {
		cassandra Cassandra
		err       error
	}
	ch := make(chan result)
	go func() {
		cluster := gocql.NewCluster(cfg.Host)
		cluster.Port = cfg.Port
		cluster.Consistency = gocql.Quorum
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
		return res.cassandra, res.err
	case <-ctx.Done():
		return Cassandra{}, ctx.Err()
	}
}

func CloseCassandra(ctx context.Context, cassandra Cassandra) error {
	cassandra.Session.Close()
	return nil
}
