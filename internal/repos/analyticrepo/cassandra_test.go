package analyticrepo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
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

func TestCassandraAnalyticRepo_CreateSuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()
	a := &domain.Analytic{
		Alias:     uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		Clicks:    0,
		Version:   0,
	}
	if err := repo.Create(ctx, a); err != nil {
		t.Errorf("expected no error, got %+v", a)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, a.Alias)
	})
}

func TestCassandraAnalyticRepo_GetsuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()
	expected := &domain.Analytic{
		Alias:     uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		Clicks:    0,
		Version:   0,
	}
	if err := repo.Create(ctx, expected); err != nil {
		t.Skip()
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, expected.Alias)
	})

	got, err := repo.Get(ctx, expected.Alias)
	if err != nil {
		if errs.IsNotFound(err) {
			t.Fatalf("expected no error, got %+v", err)
		}
		t.Skip(err)
	}

	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func TestCassandraAnalyticRepo_UpdatesuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()
	expected := &domain.Analytic{
		Alias:     uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		Clicks:    0,
		Version:   0,
	}
	if err := repo.Create(ctx, expected); err != nil {
		t.Skip()
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, expected.Alias)
	})

	expected.Clicks += 1
	expected.UpdatedAt = time.Now().UTC()
	if err := repo.Update(ctx, expected); err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}

	got, err := repo.Get(ctx, expected.Alias)
	if err != nil {
		if errs.IsNotFound(err) {
			t.Fatalf("expected no error, got %+v", err)
		}
		t.Skip(err)
	}

	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func TestCassandraAnalyticRepo_DeletesuccessForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()
	expected := &domain.Analytic{
		Alias:     uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		Clicks:    0,
		Version:   0,
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
	}
}
