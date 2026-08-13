package ports

import "context"

type Migrator interface {
	Execute(ctx context.Context, query string) error
}
