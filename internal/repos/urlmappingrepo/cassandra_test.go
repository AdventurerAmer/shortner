package urlmappingrepo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
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

var testContext cassandraTestContext

func TestMain(m *testing.M) {
	dir, _ := os.Getwd()
	fmt.Printf("working dir: %+v\n", dir)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %+v\n", err)
		os.Exit(1)
	}

	testContext.logger = logging.New(cfg)

	testContext.cassandra, err = infra.ConnectToCassandra(context.TODO(), &cfg.Infrastructure.Database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to cassandra failed: %+v", err)
		os.Exit(1)
	}
	testContext.keyspace = cfg.Infrastructure.Database.Keyspace

	exitCode := m.Run()

	infra.CloseCassandra(context.TODO(), testContext.cassandra)

	os.Exit(exitCode)
}

func TestCassandraURLMappingRepo_CreateSuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := &domain.URLMapping{
		ShortURL:  uuid.NewString(),
		LongURL:   "www.example.com/examples",
		CreatedAt: time.Now().UTC(),
		UserId:    uuid.NewString(),
	}
	if err := repo.Create(ctx, m); err != nil {
		t.Errorf("expected no error, got %+v", m)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, m.ShortURL)
	})
}

func TestCassandraURLMappingRepo_GetsuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := &domain.URLMapping{
		ShortURL:  uuid.NewString(),
		LongURL:   "www.example.com/examples",
		CreatedAt: time.Now().UTC(),
		UserId:    uuid.NewString(),
	}
	if err := repo.Create(ctx, expected); err != nil {
		t.Skip()
	}

	got, err := repo.Get(ctx, expected.ShortURL)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, expected.ShortURL)
	})

	if cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func TestCassandraURLMappingRepo_DeletesuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := &domain.URLMapping{
		ShortURL:  uuid.NewString(),
		LongURL:   "www.example.com/examples",
		CreatedAt: time.Now().UTC(),
		UserId:    uuid.NewString(),
	}
	if err := repo.Create(ctx, expected); err != nil {
		t.Skip()
	}

	err := repo.Delete(ctx, expected.ShortURL)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
}

func createRepo(t *testing.T) *cassandraRepo {
	t.Helper()
	return &cassandraRepo{
		session:  testContext.cassandra.Session,
		keyspace: testContext.keyspace,
	}
}
