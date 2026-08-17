package urlmapping

import (
	"context"
	"testing"
	"time"

	cassandraMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/cassandra/v1"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

func TestCassandraURLMappingRepo(t *testing.T) {
	t.Parallel()

	if err := test.ChangeToRootDir(); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cassandra := test.Cassandra(ctx, t)
	defer cassandra.Disconnect(context.Background())

	if err := cassandraMigratorV1.Run(ctx, &cassandra); err != nil {
		t.Fatal(err)
	}

	keyspace := cfg.Infrastructure.Cassandra.Keyspace
	repo := NewCassandra(cassandra.Session, keyspace, ports.NewCacheStub())

	exampleURL := "www.example.com/examples"

	t.Run("CreateSucceedsForValidInput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		expected := &domain.URLMapping{
			UserId:    uuid.NewString(),
			CreatedAt: time.Now().UTC(),
			Alias:     uuid.NewString(),
			LongURL:   exampleURL,
		}
		if err := repo.Create(ctx, expected); err != nil {
			t.Fatalf("expected no error, got %+v", nil)
		}
		got, err := repo.Get(ctx, expected.Alias)
		if err != nil {
			if errs.IsNotFound(err) {
				t.Fatalf("expected no error, got %+v", err)
			} else {
				t.Skipf("failed to get url mapping: %+v", err)
			}
		}

		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	})

	t.Run("GetSucceedsForValidInput", func(t *testing.T) {
		expected := &domain.URLMapping{
			UserId:    uuid.NewString(),
			CreatedAt: time.Now().UTC(),
			Alias:     uuid.NewString(),
			LongURL:   exampleURL,
		}
		if err := repo.Create(ctx, expected); err != nil {
			t.Skipf("failed to create url mapping: %+v", err)
		}

		got, err := repo.Get(ctx, expected.Alias)
		if err != nil {
			t.Fatalf("expected no error, got %+v", err)
		}

		if !cmp.Equal(expected, got, cmpopts.EquateApproxTime(time.Second)) {
			t.Errorf("expected %+v, got %+v, diff %+v", expected, got, cmp.Diff(expected, got))
		}
	})

	t.Run("DeleteSucceedsForValidInput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		expected := &domain.URLMapping{
			UserId:    uuid.NewString(),
			CreatedAt: time.Now().UTC(),
			Alias:     uuid.NewString(),
			LongURL:   exampleURL,
		}
		if err := repo.Create(ctx, expected); err != nil {
			t.Skipf("failed to create url mapping: %+v", err)
		}

		if err := repo.Delete(ctx, expected.Alias); err != nil {
			t.Fatalf("expected no error, got %+v", err)
		}

		_, err := repo.Get(ctx, expected.Alias)
		if err == nil || !errs.IsNotFound(err) {
			t.Fatalf("expected a not found error, got %+v", err)
		}
	})
}
