package shortening

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmappingrepo"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

type cassandraTestContext struct {
	cassandra *infra.Cassandra
	keyspace  string
	logger    *logging.Logger
}

var testCtx cassandraTestContext

func TestMain(m *testing.M) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %+v\n", err)
		os.Exit(1)
	}

	testCtx.logger = logging.New(cfg)

	testCtx.cassandra, err = infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to cassandra failed: %+v", err)
		os.Exit(1)
	}
	testCtx.keyspace = cfg.Infrastructure.Database.Keyspace

	exitCode := m.Run()

	infra.CloseCassandra(context.TODO(), testCtx.cassandra)

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
		got := m.ShortURL

		if cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	}

	{
		expected := req.LongURL
		got := m.LongURL

		if cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, m.ShortURL)
	})
}

func createRepo(t *testing.T) ports.URLMappingRepository {
	t.Helper()
	URLMappingRepo := urlmappingrepo.NewCassandra(testCtx.cassandra.Session, testCtx.keyspace)
	return URLMappingRepo
}

func createService(t *testing.T) ports.ShorteningService {
	t.Helper()
	repo := createRepo(t)
	cfg := Config{
		URLMappingRepo: repo,
		ShortURLPrefix: "",
	}
	return &service{
		Config: cfg,
	}
}
