package infra

import (
	"context"

	"github.com/AdventurerAmer/shortner/config"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type Cassandra struct {
	Session *gocql.Session
}

func ConnectToCassandra(ctx context.Context, cfg *config.CassandraDatabaseConfig) (Cassandra, error) {
	cluster := gocql.NewCluster(cfg.Host)
	cluster.Keyspace = cfg.Keyspace
	cluster.Port = cfg.Port
	cluster.Consistency = gocql.Quorum
	cluster.ConnectTimeout = cfg.ConnTimeout

	session, err := cluster.CreateSession()
	if err != nil {
		return Cassandra{}, err
	}

	return Cassandra{Session: session}, nil
}

func CloseCassandra(ctx context.Context, cassandra Cassandra) {
	cassandra.Session.Close()
}
