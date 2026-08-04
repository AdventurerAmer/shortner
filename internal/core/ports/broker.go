package ports

import (
	"context"
)

type Producer interface {
	Send(ctx context.Context, key string, data []byte) error
}

type ConsumerHandlerFunc = func(ctx context.Context, key string, data []byte) error

type Consumer interface {
	Receive(ctx context.Context, handler ConsumerHandlerFunc) error
}
