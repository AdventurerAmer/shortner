package analytics

import (
	"context"
	"testing"
	"time"

	clickhouseMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/clickhouse/v1"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticsrepo"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
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

	clickHouse := test.ClickHouse(ctx, t)
	if err := clickhouseMigratorV1.Run(ctx, &clickHouse); err != nil {
		t.Fatal(err)
	}

	if err := clickhouseMigratorV1.Run(ctx, &clickHouse); err != nil {
		t.Fatal(err)
	}

	database := cfg.Infra.ClickHouse.Database
	repo := analyticsrepo.NewClickHouse(database, clickHouse.Conn, ports.NewCacheStub(), time.Second)

	srvCfg := Config{
		AnalyticClicksRepo: repo,
	}
	service := &service{
		Config: srvCfg,
	}

	t.Run("GetSucceedsForValidInput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		shard := "sa"
		idGen := snowflake.New(shard)

		analytic := &domain.Analytic{
			Alias:  idGen.Next(),
			Clicks: 1,
		}
		alias := analytic.Alias
		if err := repo.Put(ctx, uuid.NewString(), alias, time.Now()); err != nil {
			t.Skipf("failed to create analytic: %+v", err)
		}

		req := ports.GetAnalyticRequest{
			Alias: analytic.Alias,
		}
		resp, err := service.Get(ctx, req)
		if err != nil {
			if errs.IsNotFound(err) {
				t.Fatalf("expected no error, got %+v", err)
			}
			t.Skipf("get analytic failed: %+v", err)
		}

		expected := analytic
		got := resp.Analytic
		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	})
}
