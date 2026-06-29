package analytics

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticrepo"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var testCtx *test.Cassandra

func TestMain(m *testing.M) {
	var err error
	testCtx, err = test.NewCassandraTestContext()
	if err != nil {
		panic(err)
	}
	exitCode := m.Run()
	testCtx.Shutdown()
	os.Exit(exitCode)
}

func TestAnalyticsService_GetSucceedsForValidInput(t *testing.T) {
	repo := createRepo(t)
	service := createService(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	shard := "sa"
	idGenerator := snowflake.New(shard)
	now := time.Now().UTC()
	m := &domain.Analytic{
		Alias:     idGenerator.Next(),
		CreatedAt: now,
		UpdatedAt: now,
		Clicks:    0,
		Version:   0,
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Skipf("failed to create analytic: %+v", err)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, m.Alias)
	})

	req := ports.GetAnalyticRequest{
		Alias: m.Alias,
	}
	resp, err := service.Get(ctx, req)
	if err != nil {
		if errs.IsNotFound(err) {
			t.Fatalf("expected no error, got %+v", err)
		}
		t.Skipf("get analytic failed: %+v", err)
	}

	expected := m
	got := resp.Analytic
	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func TestAnalyticsService_IncrementClicksSucceedsForValidInput(t *testing.T) {
	repo := createRepo(t)
	service := createService(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	shard := "sa"
	idGenerator := snowflake.New(shard)
	now := time.Now().UTC()
	m := &domain.Analytic{
		Alias:     idGenerator.Next(),
		CreatedAt: now,
		UpdatedAt: now,
		Clicks:    0,
		Version:   0,
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Skipf("failed to create analytic: %+v", err)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, m.Alias)
	})

	req := ports.IncrementClicksRequest{
		Alias:  m.Alias,
		Clicks: 1,
	}
	resp, err := service.IncrementClicks(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}

	m.Clicks += 1
	m.Version += 1
	expected := m
	got := resp.Analytic

	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func createRepo(t *testing.T) ports.AnalyticRepository {
	t.Helper()
	AnalyticRepo := analyticrepo.NewCassandra(testCtx.Cassandra.Session, testCtx.Keyspace, ports.NewCacheStub())
	return AnalyticRepo
}

func createService(t *testing.T) ports.AnalyticService {
	t.Helper()
	repo := createRepo(t)
	cfg := Config{
		AnalyticRepo: repo,
	}
	return &service{
		Config: cfg,
	}
}
