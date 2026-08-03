package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"os"
	"time"

	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/google/uuid"
)

type bucket struct {
	ch            chan string
	m             map[string]int
	lastDumpTime  time.Time
	eventProducer *ports.EventProducer
}

func (b *bucket) dump() {
	l := len(b.m)
	if l == 0 {
		return
	}

	uuids := make([]string, l)
	keys := make([]string, l)
	values := make([]int, l)

	i := 0
	for key, val := range b.m {
		uuids[i] = uuid.NewString()
		keys[i] = key
		values[i] = val
		i += 1
	}

	clear(b.m)
	b.lastDumpTime = time.Now()

	event := domain.ClicksBatchEvent{
		UUIds:   uuids,
		Aliases: keys,
		Clicks:  values,
	}

	b.eventProducer.Fire(context.Background(), event)
}

func newBucket(chCap int, mCap int, eventProducer *ports.EventProducer) *bucket {
	return &bucket{
		ch:            make(chan string, chCap),
		m:             make(map[string]int, mCap),
		eventProducer: eventProducer,
	}
}

type collector struct {
	chans []chan string
}

func newCollector(count, batchSize int, eventProducer *ports.EventProducer) *collector {
	chans := make([]chan string, count)
	for i := range count {
		b := newBucket(256, batchSize, eventProducer)
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

	groupId := "clicks-batcher"
	logger := logging.New(cfg).With(slog.String("service", groupId))

	writer := infra.NewKafkaWriter(cfg.Infrastructure.Kafka, domain.ClicksBatchTopic)
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("'producer.Close' failed", "error", err)
		}
	}()

	orch := goorch.New(context.Background())
	defer orch.CancelAndWait()

	producer := brokers.NewKafkaProducer(writer)

	eventProducer, err := ports.NewEventProducer(producer, orch)
	if err != nil {
		logger.Error("create event producer failed", "error", err)
		os.Exit(1)
	}

	bucketCount := 256
	batchSize := 1024
	collector := newCollector(bucketCount, batchSize, eventProducer)

	reader := infra.NewKafkaReader(cfg.Infrastructure.Kafka, domain.ClicksTopic, groupId)
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("'reader.Close' failed", "error", err)
		}
	}()

	consumer := brokers.NewKafkaConsumer(reader)

	h := func(key string, data []byte) {
		logger.Info("recived event", "key", key)

		var event domain.ClickEvent
		if err := json.Unmarshal(data, &event); err != nil {
			logger.Error("'json.Unmarshal' failed", "key", key, "error", err)
			return
		}

		alias := event.Alias
		collector.inc(alias)
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		logger.Error("'consumer.Receive' failed", "error", err)
	}
}
