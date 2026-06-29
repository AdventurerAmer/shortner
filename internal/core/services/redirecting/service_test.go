package redirecting

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmappingrepo"
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

func TestRedirectingService_RedirectSuccessForValidInput(t *testing.T) {
	repo := createRepo(t)
	service := createService(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	shard := "sa"
	idGenerator := snowflake.New(shard)

	m := &domain.URLMapping{
		Alias:     idGenerator.Next(),
		LongURL:   "www.example.com/examples",
		CreatedAt: time.Now().UTC(),
		UserId:    uuid.NewString(),
	}

	if err := repo.Create(ctx, m); err != nil {
		t.Skipf("failed to create url mapping: %+v", err)
	}

	req := ports.RedirectRequest{
		Alias: m.Alias,
	}
	resp, err := service.Redirect(ctx, req)
	if err != nil {
		t.Errorf("expected no error, got %+v", err)
	}

	expected := m.LongURL
	got := resp.LongURL

	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, m.Alias)
	})
}

func createRepo(t *testing.T) ports.URLMappingRepository {
	t.Helper()
	URLMappingRepo := urlmappingrepo.NewCassandra(testCtx.Cassandra.Session, testCtx.Keyspace, ports.NewCacheStub(), testCtx.Logger)
	return URLMappingRepo
}

func createService(t *testing.T) ports.RedirectingService {
	t.Helper()
	repo := createRepo(t)
	cfg := Config{
		URLMappingRepo: repo,
	}
	return &service{
		Config: cfg,
		logger: testCtx.Logger,
	}
}
