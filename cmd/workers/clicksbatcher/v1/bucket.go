package v1

import (
	"context"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/google/uuid"
)

type bucket struct {
	ch            chan string
	m             map[string]int
	lastDumpTime  time.Time
	eventProducer *ports.EventProducer
}

func newBucket(chCap int, mCap int, eventProducer *ports.EventProducer) *bucket {
	return &bucket{
		ch:            make(chan string, chCap),
		m:             make(map[string]int, mCap),
		eventProducer: eventProducer,
	}
}

func (b *bucket) dump() {
	l := len(b.m)
	if l == 0 {
		return
	}

	uuids := make([]string, l)
	keys := make([]string, l)
	values := make([]int, l)

	idx := 0
	for key, val := range b.m {
		uuids[idx] = uuid.NewString()
		keys[idx] = key
		values[idx] = val
		idx += 1
	}

	event := domain.ClicksBatchEvent{
		UUIds:   uuids,
		Aliases: keys,
		Clicks:  values,
	}
	b.eventProducer.Fire(context.Background(), event)

	clear(b.m)
	b.lastDumpTime = time.Now()
}

func (b *bucket) collect(ctx context.Context, batchSize int) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			b.dump()
			return
		case key, ok := <-b.ch:
			if !ok {
				return
			}
			b.m[key] += 1
			if len(b.m) == batchSize {
				b.dump()
			}
		case <-t.C:
			if time.Since(b.lastDumpTime) >= time.Second {
				b.dump()
			}
		}
	}
}
