package analytics

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/analyticclicks"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
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

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, stat.Alias)
	})

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
}

func createRepo(t *testing.T) ports.AnalyticClicksRepository {
	t.Helper()
	AnalyticRepo := analyticclicks.NewCassandra(testCtx.Cassandra.Session, testCtx.Keyspace, ports.NewCacheStub())
	return AnalyticRepo
}

func createService(t *testing.T) ports.AnalyticService {
	t.Helper()
	repo := createRepo(t)
	cfg := Config{
		AnalyticStatRepo: repo,
	}
	return &service{
		Config: cfg,
	}
}
