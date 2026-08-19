package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
)

type Cache interface {
	Get(ctx context.Context, key string, v any) error
	Put(ctx context.Context, key string, v any, TTL time.Duration) error
}

type cacheStub struct{}

func (c *cacheStub) Get(ctx context.Context, key string, v any) error {
	return errs.NewNotFound(nil, "key not found")
}

func (c *cacheStub) Put(ctx context.Context, key string, v any, TTL time.Duration) error {
	return nil
}

func NewCacheStub() Cache {
	return &cacheStub{}
}
