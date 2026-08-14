package test

import (
	"context"
	"testing"

	"github.com/AdventurerAmer/shortner/infra"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/testcontainers/testcontainers-go"

	testClickhouse "github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func ClickHouse(ctx context.Context, t *testing.T) infra.ClickHouse {
	t.Helper()

	user, password, dbname := "clickhouse", "password", "default"
	ctr, err := testClickhouse.Run(ctx, "clickhouse/clickhouse-server:25.8",
		testClickhouse.WithUsername(user),
		testClickhouse.WithPassword(password),
		testClickhouse.WithDatabase(dbname),
	)
	testcontainers.CleanupContainer(t, ctr)

	connStr, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	options, err := clickhouse.ParseDSN(connStr)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		t.Fatal(err)
	}

	if err := conn.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	clickHouse := infra.ClickHouse{
		Conn: conn,
	}
	return clickHouse
}
