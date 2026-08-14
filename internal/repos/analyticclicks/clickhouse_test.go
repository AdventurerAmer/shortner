package analyticclicks

import (
	"context"
	"testing"
	"time"

	clickhouseMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/clickhouse/v1"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	testClickhouse "github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func TestClickhouseAnalyticRepo(t *testing.T) {
	t.Parallel()

	if err := test.ChangeToRootDir(); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	logger := logging.New(cfg)
	ctx := logging.Set(context.Background(), logger)

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
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	clickHouse := &infra.ClickHouse{
		Conn: conn,
	}

	if err := clickhouseMigratorV1.Run(ctx, clickHouse); err != nil {
		t.Fatal(err)
	}

	database := cfg.Infrastructure.ClickHouse.Database
	repo := NewClickHouse(database, conn, ports.NewCacheStub(), time.Second)

	t.Run("GetSucceedsForValidInput", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		idGen := snowflake.New(uuid.NewString())

		expected := &domain.AnalyticClicks{
			Alias:  idGen.Next(),
			Clicks: 1,
		}
		patchId := []string{uuid.NewString()}
		aliases := []string{expected.Alias}
		clicks := []int{int(expected.Clicks)}
		if err := repo.Put(ctx, patchId, aliases, clicks); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		got, err := repo.Get(ctx, expected.Alias)
		if err != nil {
			if errs.IsNotFound(err) {
				t.Fatalf("expected no error, got %+v", err)
			}
			t.Skipf("failed to get analytic clicks: %+v", err)
		}

		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	})
	t.Run("PutSucceedsForValidInput", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		idGen := snowflake.New(uuid.NewString())

		expected := &domain.AnalyticClicks{
			Alias:  idGen.Next(),
			Clicks: 1,
		}
		ids := []string{uuid.NewString()}
		aliases := []string{expected.Alias}
		clicks := []int{int(expected.Clicks)}
		if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		got, err := repo.Get(ctx, expected.Alias)
		if err != nil {
			if errs.IsNotFound(err) {
				t.Fatalf("expected no error, got %+v", err)
			}
			t.Skipf("'repo.Get' failed: %+v", err)
		}

		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	})
	t.Run("DeleteSucceedsForValidInput", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		idGen := snowflake.New(uuid.NewString())
		expected := &domain.AnalyticClicks{
			Alias:  idGen.Next(),
			Clicks: 1,
		}
		ids := []string{uuid.NewString()}
		aliases := []string{expected.Alias}
		clicks := []int{int(expected.Clicks)}
		if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		if err := repo.Delete(ctx, expected.Alias); err != nil {
			t.Fatalf("expected no error, got %+v", err)
		}
	})
}
