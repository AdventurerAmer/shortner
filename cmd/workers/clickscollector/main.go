package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/sony/gobreaker/v2"
)

var analyticsCB = gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
	Name:        "analytics",
	Timeout:     30 * time.Second, // Time in Open state before Half-Open
	MaxRequests: 5,                // Requests allowed in Half-Open
	Interval:    60 * time.Second, // Clear counts periodically in Closed
	ReadyToTrip: func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures > 5
	},
	IsSuccessful: func(err error) bool {
		return err == nil
	},
})

type bucket struct {
	ch           chan string
	m            map[string]int
	lastDumpTime time.Time
	producer     ports.Producer
}

func (b *bucket) dump() {
	l := len(b.m)
	if l == 0 {
		return
	}

	keys := make([]string, l)
	values := make([]int, l)

	i := 0
	for key, val := range b.m {
		keys[i] = key
		values[i] = val
		i += 1
	}

	clear(b.m)
	b.lastDumpTime = time.Now()

	go func(ctx context.Context) {
		event := domain.ClicksBatchEvent{
			UUId:    uuid.NewString(),
			Aliases: keys,
			Clicks:  values,
		}

		retryFunc := func() error {

			_, err := analyticsCB.Execute(func() ([]byte, error) {
				slog.Debug("sending batched clicks event started")

				dctx, cancel := context.WithTimeout(ctx, domain.ClicksBatchEventTimeout)
				defer cancel()

				key := event.UUId
				data, err := json.Marshal(&event)
				if err != nil {
					return nil, fmt.Errorf("'json.Marshal' failed: %w", err)
				}

				if err := b.producer.Send(dctx, key, data); err != nil {
					return nil, fmt.Errorf("'producer.Send' failed: %w", err)
				}

				logger := logging.Get(ctx)
				logger.Debug("sending batched clicks event succeeded", "event", event)

				return nil, nil
			})

			return err
		}

		if err := retry.Do(
			retryFunc,
			retry.Attempts(10),
			retry.Delay(300*time.Millisecond),
			retry.MaxJitter(200*time.Millisecond),
			retry.DelayType(retry.BackOffDelay),
			retry.LastErrorOnly(true),
			retry.Context(ctx),
		); err != nil {
			logger := logging.Get(ctx)
			logger.Error("sending clicks event failed", "event", event, "error", err)
		}
	}(context.Background())
}

func newBucket(chCap int, mCap int, producer ports.Producer) *bucket {
	return &bucket{
		ch:       make(chan string, chCap),
		m:        make(map[string]int, mCap),
		producer: producer,
	}
}

type collector struct {
	chans []chan string
}

func newCollector(count, batchSize int, producer ports.Producer) *collector {
	chans := make([]chan string, count)
	for i := range count {
		b := newBucket(256, batchSize, producer)
		chans[i] = b.ch

		go func(b *bucket) {
			t := time.NewTicker(time.Second)
			defer t.Stop()
			for {
				select {
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
		}(b)
	}
	return &collector{
		chans: chans,
	}
}

func (c *collector) inc(key string) {
	h := fnv.New32a()
	h.Write([]byte(key))
	index := int(h.Sum32()) % len(c.chans)
	c.chans[index] <- key
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	groupId := "clicks-collector"
	logger := logging.New(cfg).With(slog.String("service", groupId))

	writer := infra.NewKafkaWriter(cfg.Infrastructure.Kafka, domain.ClicksBatchTopic)
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("'producer.Close' failed", "error", err)
		}
	}()

	producer := brokers.NewKafkaProducer(writer)

	bucketCount := 256
	batchSize := 1024
	collector := newCollector(bucketCount, batchSize, producer)

	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.ClicksTopic, groupId)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)

	h := func(key string, data []byte) {
		logger.Info("recived event", "status", "started", "key", key)
		defer logger.Info("recived event", "status", "ended", "key", key)

		var event domain.ClickEvent
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("'json.Unmarshal' failed", "key", key, "error", err)
			return
		}

		alias := event.Alias
		logger.Debug("event alias", "alias", alias)
		collector.inc(alias)
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		slog.Info("'consumer.Receive' failed", "error", err)
	}
}
