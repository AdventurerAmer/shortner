package v1

import (
	"context"

	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/cespare/xxhash/v2"
)

type collector struct {
	orch   *goorch.Orchestrator
	keysCh []chan string
}

func newCollector(bucketCount, bucketCapacity, batchSize int, orch *goorch.Orchestrator, eventProducer *ports.EventProducer) *collector {
	chans := make([]chan string, bucketCount)
	for idx := range bucketCount {
		b := newBucket(bucketCapacity, batchSize, eventProducer)
		orch.Go(context.Background(), func(ctx context.Context) {
			b.collect(ctx, batchSize)
		})
		chans[idx] = b.ch
	}
	return &collector{
		keysCh: chans,
	}
}

func (c *collector) inc(key string) {
	hash := xxhash.Sum64([]byte(key))
	idx := hash % uint64(len(c.keysCh))
	c.keysCh[idx] <- key
}
