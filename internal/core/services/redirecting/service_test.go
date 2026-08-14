package redirecting

import (
	"context"
	"testing"
	"time"

	cassandraMigratorV1 "github.com/AdventurerAmer/shortner/cmd/migrators/cassandra/v1"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmapping"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

func TestRedirectingService_CassandraRepo(t *testing.T) {
	if err := test.ChangeToRootDir(); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cassandra := test.Cassandra(ctx, t)
	defer infra.CloseCassandra(context.Background(), cassandra)

	if err := cassandraMigratorV1.Run(ctx, &cassandra); err != nil {
		t.Fatal(err)
	}

	keyspace := cfg.Infrastructure.Cassandra.Keyspace
	repo := urlmapping.NewCassandra(cassandra.Session, keyspace, ports.NewCacheStub())

	srvCfg := Config{
		URLMappingRepo: repo,
	}
	service := &service{
		Config: srvCfg,
	}
	t.Run("RedirectSucceedsForValidInput", func(t *testing.T) {
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
	})
}
