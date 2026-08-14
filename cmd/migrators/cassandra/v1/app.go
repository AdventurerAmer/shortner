package v1

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/apps/migrator"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
)

func Run(ctx context.Context, m ports.Migrator) error {
	logger := logging.Get(ctx)

	glob := "internal/migrations/cassandra/*.cql"
	app, err := migrator.New(logger, m)
	if err != nil {
		return fmt.Errorf("'migrator.New' failed: %w", err)
	}
	if err := app.Run(ctx, glob); err != nil {
		return fmt.Errorf("'app.Run' failed: %w", err)
	}

	return nil
}
