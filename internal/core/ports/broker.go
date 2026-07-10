package ports

import "context"

type Producer interface {
	Send(ctx context.Context, key string, data []byte) error
}

type ConsumerHandlerFunc = func(key string, data []byte)

type Consumer interface {
	Receive(ctx context.Context, handler ConsumerHandlerFunc) error
}
