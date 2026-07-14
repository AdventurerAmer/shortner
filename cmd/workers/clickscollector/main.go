package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/infra"
	"github.com/AdventurerAmer/shortner/internal/brokers"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/logging"
)

type collector struct {
	mu sync.RWMutex
	m  map[string]int
}

func newCollector() *collector {
	return &collector{
		m: make(map[string]int),
	}
}

func (c *collector) inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] += 1
}

func (c *collector) grab() (string, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var (
		key string
		val int
	)
	for k, v := range c.m {
		key = k
		val = v
		break
	}
	if key != "" {
		delete(c.m, key)
	}
	return key, val
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %+v\n", err)
		os.Exit(1)
	}

	groupId := "clicks-collector"
	logger := logging.New(cfg).With(slog.String("service", groupId))

	collector := newCollector()

	writer := infra.NewKafkaWriter(cfg.Infrastructure.Kafka, domain.CollectedClicksTopic)
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("'producer.Close' failed", "error", err)
		}
	}()

	producer := brokers.NewKafkaProducer(writer)

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for range t.C {
			func() {
				alias, clicks := collector.grab()
				ctx, cancel := context.WithTimeout(context.Background(), domain.ClickEventTimeout)
				defer cancel()
				event := domain.CollectedClicksEvent{
					Alias:  alias,
					Clicks: clicks,
				}
				data, err := json.Marshal(event)
				if err != nil {
					return
				}
				if err := producer.Send(ctx, alias, data); err != nil {
					return
				}
			}()
		}
	}()

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
		collector.inc(alias)
	}
	if err := consumer.Receive(context.Background(), h); err != nil {
		slog.Info("'consumer.Receive' failed", "error", err)
	}
}
