package urlmappingrepo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
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

func TestCassandraURLMappingRepo_CreateSuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := &domain.URLMapping{
		Alias:     uuid.NewString(),
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
		repo.Delete(dctx, m.Alias)
	})
}

func TestCassandraURLMappingRepo_GetsuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := &domain.URLMapping{
		Alias:     uuid.NewString(),
		LongURL:   "www.example.com/examples",
		CreatedAt: time.Now().UTC(),
		UserId:    uuid.NewString(),
	}
	if err := repo.Create(ctx, expected); err != nil {
		t.Skip()
	}

	got, err := repo.Get(ctx, expected.Alias)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, expected.Alias)
	})

	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func TestCassandraURLMappingRepo_DeletesuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := &domain.URLMapping{
		Alias:     uuid.NewString(),
		LongURL:   "www.example.com/examples",
		CreatedAt: time.Now().UTC(),
		UserId:    uuid.NewString(),
	}
	if err := repo.Create(ctx, expected); err != nil {
		t.Skip()
	}

	err := repo.Delete(ctx, expected.Alias)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
}

func createRepo(t *testing.T) *cassandraRepo {
	t.Helper()
	return &cassandraRepo{
		session:  testCtx.Cassandra.Session,
		keyspace: testCtx.Keyspace,
		cache:    ports.NewCacheStub(),
		logger:   testCtx.Logger,
	}
}
