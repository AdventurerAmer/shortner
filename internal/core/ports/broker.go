package ports

import "context"

type Producer interface {
	Send(ctx context.Context, key string, data []byte) error
}
