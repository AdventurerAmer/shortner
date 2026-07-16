package shortening

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	urlmappingrepo "github.com/AdventurerAmer/shortner/internal/repos/urlmapping"
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

func TestShorteningService_ShortenSucceedsForValidInput(t *testing.T) {
	repo := createRepo(t)
	service := createService(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := ports.ShortenURLRequest{
		LongURL: "https://www.example.com/examples/1234",
	}
	userId := uuid.NewString()
	resp, err := service.Shorten(ctx, userId, req)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}

	parsedURL, err := url.Parse(resp.ShortURL)
	if err != nil {
		t.Skipf("parse URL failed: %+v", err)
	}
	parts := strings.Split(parsedURL.Path, "/")
	alias := parts[len(parts)-1]

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, alias)
	})

	m, err := repo.Get(ctx, alias)
	if err != nil {
		t.Skipf("failed to get url mapping: %+v", err)
	}

	{
		expected := alias
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
}

func createRepo(t *testing.T) ports.URLMappingRepository {
	t.Helper()
	URLMappingRepo := urlmappingrepo.NewCassandra(testCtx.Cassandra.Session, testCtx.Keyspace, ports.NewCacheStub())
	return URLMappingRepo
}

func createService(t *testing.T) ports.ShorteningService {
	t.Helper()
	repo := createRepo(t)
	cfg := Config{
		URLMappingRepo: repo,
		ShortURLPrefix: "",
		IdGenerator:    snowflake.New("sa"),
	}
	return &service{
		Config: cfg,
	}
}
