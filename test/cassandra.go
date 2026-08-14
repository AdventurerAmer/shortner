package test

import (
	"context"
	"testing"

	"github.com/AdventurerAmer/shortner/infra"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func Cassandra(ctx context.Context, t *testing.T) infra.Cassandra {
	t.Helper()

	ctr, err := cassandra.Run(ctx, "cassandra:5.0.8")
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, ctr)
	host, err := ctr.ConnectionHost(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cluster := gocql.NewCluster(host)
	session, err := cluster.CreateSession()
	if err != nil {
		t.Fatal(err)
	}

	cassandra := infra.Cassandra{
		Session: session,
	}

	return cassandra
}
