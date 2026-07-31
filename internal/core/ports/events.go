package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/sony/gobreaker/v2"
)

type EventProducer struct {
	producer      Producer
	cb            *gobreaker.CircuitBreaker[[]byte]
	retryAttempts int
	timeout       time.Duration
}

func NewEventProducer(producer Producer, cb *gobreaker.CircuitBreaker[[]byte], retryAttempts int, timeout time.Duration) *EventProducer {
	return &EventProducer{
		producer:      producer,
		cb:            cb,
		retryAttempts: retryAttempts,
		timeout:       timeout,
	}
}

func (ep *EventProducer) Fire(ctx context.Context, event any) {
	retryHandler := func() error {
		key := uuid.NewString()
		data, err := json.Marshal(&event)
		if err != nil {
			return fmt.Errorf("'json.Marshal' failed: %w", err)
		}
		cbHandler := func() ([]byte, error) {
			logger := logging.Get(ctx)
			logger.Debug("sending event attempt started", "event", event)
			defer logger.Debug("sending event attempt ended", "event", event)

			dctx, cancel := context.WithTimeout(ctx, ep.timeout)
			defer cancel()

			if err := ep.producer.Send(dctx, key, data); err != nil {
				return nil, fmt.Errorf("'producer.Send' failed: %w", err)
			}

			logger.Debug("sending event succeeded", "event", event)
			return nil, nil
		}
		if _, err := ep.cb.Execute(cbHandler); err != nil {
			return fmt.Errorf("'cb.Execute' failed: %w", err)
		}
		return nil
	}
	if err := retry.Do(
		retryHandler,
		retry.Attempts(uint(ep.retryAttempts)),
		retry.MaxJitter(200*time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
	); err != nil {
		logger := logging.Get(ctx)
		logger.Error("sending clicks event failed", "event", event, "error", err)
	}
}

func DefaultEventProducer(producer Producer) *EventProducer {
	return &EventProducer{
		producer:      producer,
		cb:            domain.KafkaCB,
		retryAttempts: domain.KafkaRetryAttempts,
		timeout:       domain.KafkaTimeout,
	}
}
