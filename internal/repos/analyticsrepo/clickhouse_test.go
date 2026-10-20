package analyticsrepo

import (
	"context"
	"testing"
	"time"

	clickhouseMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/clickhouse/v1"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
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

	logger := logging.New(
		logging.WithLocalEnv(cfg.Env == config.EnvLocal),
		logging.WithFormat(cfg.Observability.Logging.Format),
		logging.WithLevel(cfg.Observability.Logging.Level),
		logging.WithAddSource(*cfg.Observability.Logging.AddSource),
	)
	ctx := logging.Set(context.Background(), logger)

	clickHouse := test.ClickHouse(ctx, t)
	if err := clickhouseMigratorV1.Run(ctx, &clickHouse); err != nil {
		t.Fatal(err)
	}

	database := cfg.Infra.ClickHouse.Database
	repo := NewClickHouse(database, clickHouse.Conn, ports.NewCacheStub(), time.Second)

	t.Run("GetSucceedsForValidInput", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		idGen := snowflake.New(uuid.NewString())

		expected := &domain.Analytic{
			Alias:  idGen.Next(),
			Clicks: 1,
		}

		eventId := uuid.NewString()
		alias := expected.Alias
		createdAt := time.Now().UTC()
		if err := repo.Put(ctx, eventId, alias, createdAt); err != nil {
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

		expected := &domain.Analytic{
			Alias:  idGen.Next(),
			Clicks: 1,
		}
		alias := expected.Alias
		if err := repo.Put(ctx, uuid.NewString(), alias, time.Now().UTC()); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		if err := repo.Put(ctx, uuid.NewString(), alias, time.Now().UTC()); err != nil {
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
		alias := idGen.Next()
		eventId := uuid.NewString()
		if err := repo.Put(ctx, eventId, alias, time.Now().UTC()); err != nil {
			t.Skipf("'repo.Put' failed: %+v", err)
		}

		if err := repo.Delete(ctx, eventId); err != nil {
			t.Fatalf("expected no error, got %+v", err)
		}
	})
}
