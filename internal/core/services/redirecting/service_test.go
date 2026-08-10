package redirecting

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/repos/urlmapping"
	"github.com/AdventurerAmer/shortner/snowflake"
	"github.com/AdventurerAmer/shortner/test"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

// var testCtx *test.Cassandra

// func TestMain(m *testing.M) {
// 	var err error
// 	testCtx, err = test.NewCassandraTestContext()
// 	if err != nil {
// 		panic(err)
// 	}
// 	exitCode := m.Run()
// 	testCtx.Shutdown()
// 	os.Exit(exitCode)
// }

func TestRedirectingService_CassandraRepo(t *testing.T) {
	if err := test.ChangeToRootDir(); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctr, err := cassandra.Run(ctx,
		"cassandra:5.0.8",
		// TODO: hardcoding migrations files for now...
		cassandra.WithInitScripts(
			filepath.Join("internal", "migrations", "cassandra", "001_create_shortner_keyspace.cql"),
			filepath.Join("internal", "migrations", "cassandra", "002_create_url_mapping_table.cql"),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, ctr)
	host, err := ctr.ConnectionHost(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cluster := gocql.NewCluster(host)
	session, err := cluster.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	keyspace := cfg.Infrastructure.Cassandra.Keyspace
	repo := urlmapping.NewCassandra(session, keyspace, ports.NewCacheStub())

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
