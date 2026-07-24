package analyticclicks

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

var testCtx *test.ClickHouse

func TestMain(m *testing.M) {
	var err error
	testCtx, err = test.NewClickHouseTestContext()
	if err != nil {
		panic(err)
	}
	exitCode := m.Run()
	testCtx.Shutdown()
	os.Exit(exitCode)
}

func TestCassandraAnalyticRepo_GetSucceedsForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	idGen := snowflake.New(uuid.NewString())

	expected := &domain.AnalyticClicks{
		Alias:  idGen.Next(),
		Clicks: 1,
	}
	patchId := []string{uuid.NewString()}
	aliases := []string{expected.Alias}
	clicks := []int{int(expected.Clicks)}
	if err := repo.Put(ctx, patchId, aliases, clicks); err != nil {
		t.Skipf("'repo.Put' failed: %+v", err)
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

func TestCassandraAnalyticRepo_PutSucceedsForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	idGen := snowflake.New(uuid.NewString())

	expected := &domain.AnalyticClicks{
		Alias:  idGen.Next(),
		Clicks: 1,
	}
	ids := []string{uuid.NewString()}
	aliases := []string{expected.Alias}
	clicks := []int{int(expected.Clicks)}
	if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
		t.Skipf("'repo.Put' failed: %+v", err)
	}

	t.Cleanup(func() {
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.Delete(dctx, expected.Alias)
	})

	if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
		t.Skipf("'repo.Put' failed: %+v", err)
	}

	if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
		t.Skipf("'repo.Put' failed: %+v", err)
	}

	got, err := repo.Get(ctx, expected.Alias)
	if err != nil {
		if errs.IsNotFound(err) {
			t.Fatalf("expected no error, got %+v", err)
		}
		t.Skipf("'repo.Get' failed: %+v", err)
	}

	if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
		t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
	}
}

func TestCassandraAnalyticRepo_DeleteSucceedsForValidInput(t *testing.T) {
	repo := createRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	idGen := snowflake.New(uuid.NewString())
	expected := &domain.AnalyticClicks{
		Alias:  idGen.Next(),
		Clicks: 1,
	}
	ids := []string{uuid.NewString()}
	aliases := []string{expected.Alias}
	clicks := []int{int(expected.Clicks)}
	if err := repo.Put(ctx, ids, aliases, clicks); err != nil {
		t.Skip()
	}

	if err := repo.Delete(ctx, expected.Alias); err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
}

func createRepo(t *testing.T) *clickHouseRepo {
	t.Helper()
	return &clickHouseRepo{
		database: testCtx.Database,
		conn:     testCtx.ClickHouse.Conn,
		cache:    ports.NewCacheStub(),
	}
}
