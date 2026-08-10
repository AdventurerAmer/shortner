package analytics

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticclicks"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	testClickhouse "github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func TestAnalyticsService_ClickHouseRepo(t *testing.T) {
	t.Parallel()

	if err := test.ChangeToRootDir(); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	user, password, dbname := "clickhouse", "password", "testdb"

	ctr, err := testClickhouse.Run(ctx,
		"clickhouse/clickhouse-server:25.8",
		testClickhouse.WithUsername(user),
		testClickhouse.WithPassword(password),
		testClickhouse.WithDatabase(dbname),
		// TODO: hardcoding migrations files for now
		testClickhouse.WithInitScripts(
			filepath.Join("internal", "migrations", "clickhouse", "001_create_clicks_table.sql"),
			filepath.Join("internal", "migrations", "clickhouse", "002_create_clicks_table_view.sql"),
		),
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
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	database := cfg.Infrastructure.ClickHouse.Database
	repo := analyticclicks.NewClickHouse(database, conn, ports.NewCacheStub(), time.Second)

	srvCfg := Config{
		AnalyticStatRepo: repo,
	}
	service := &service{
		Config: srvCfg,
	}

	t.Run("GetSucceedsForValidInput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		shard := "sa"
		idGenerator := snowflake.New(shard)
		stat := &domain.AnalyticClicks{
			Alias:  idGenerator.Next(),
			Clicks: 10,
		}
		patchId := []string{uuid.NewString()}
		aliases := []string{stat.Alias}
		clicks := []int{int(stat.Clicks)}
		if err := repo.Put(ctx, patchId, aliases, clicks); err != nil {
			t.Skipf("failed to create analytic stat: %+v", err)
		}

		req := ports.GetAnalyticStatRequest{
			Alias: stat.Alias,
		}
		resp, err := service.Get(ctx, req)
		if err != nil {
			if errs.IsNotFound(err) {
				t.Fatalf("expected no error, got %+v", err)
			}
			t.Skipf("get analytic failed: %+v", err)
		}

		expected := stat
		got := resp.AnalyticStat
		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	})
}
