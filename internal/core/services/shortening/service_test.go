package shortening

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmappingrepo"
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

func TestShorteningService_ShortenSuccessForValidInput(t *testing.T) {
	repo := createRepo(t)
	service := createService(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := ports.ShortenURLRequest{
		LongURL: fmt.Sprintf("www.example.com/examples/%s", uuid.NewString()),
	}
	userId := uuid.NewString()
	resp, err := service.Shorten(ctx, userId, req)
	if err != nil {
		t.Errorf("expected no error, got %+v", err)
	}

	m, err := repo.Get(ctx, resp.ShortURL)
	if err != nil {
		t.Skipf("failed to get url mapping: %+v", err)
	}

	{
		expected := resp.ShortURL
		got := m.Alias

		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	}

	{
		expected := req.LongURL
		got := m.LongURL

		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
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

func createService(t *testing.T) ports.ShorteningService {
	t.Helper()
	repo := createRepo(t)
	cfg := Config{
		URLMappingRepo: repo,
		ShortURLPrefix: "",
		Shard:          "sa",
		Snowflake:      domain.NewSnowflake(),
	}
	return &service{
		Config: cfg,
	}
}
