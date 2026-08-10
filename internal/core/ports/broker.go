package ports

import (
	"context"
)

type Producer interface {
	Send(ctx context.Context, key string, data []byte) error
}

type ConsumerMessage struct {
	Key         string
	Data        []byte
	OriginalMsg any
}

type ConsumerHandlerFunc = func(ctx context.Context, msg ConsumerMessage) error

type Consumer interface {
	Receive(ctx context.Context) <-chan ConsumerMessage
	Ack(ctx context.Context, msg ConsumerMessage) error
}
