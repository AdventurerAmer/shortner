package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/async/goorch"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/sony/gobreaker/v2"
)

type EventProducerOption = func(ep *EventProducer) error

func WithCircuitBreaker(cb *gobreaker.CircuitBreaker[[]byte]) EventProducerOption {
	return func(ep *EventProducer) error {
		if cb == nil {
			return fmt.Errorf("'cb' is nil")
		}
		ep.cb = cb
		return nil
	}
}

func WithRetryAttempts(attempts int) EventProducerOption {
	return func(ep *EventProducer) error {
		if attempts <= 0 {
			return fmt.Errorf("'attempts' is not positive")
		}
		ep.retryAttempts = attempts
		return nil
	}
}

func WithTimeout(timeout time.Duration) EventProducerOption {
	return func(ep *EventProducer) error {
		if timeout == 0 {
			return fmt.Errorf("'timeout' is zero")
		}
		ep.timeout = timeout
		return nil
	}
}

type EventProducer struct {
	producer      Producer
	orch          *goorch.Orchestrator
	cb            *gobreaker.CircuitBreaker[[]byte]
	retryAttempts int
	timeout       time.Duration
}

func NewEventProducer(producer Producer, orch *goorch.Orchestrator, opts ...EventProducerOption) (*EventProducer, error) {
	ep := &EventProducer{
		producer:      producer,
		orch:          orch,
		cb:            domain.SendEventCircuitBreaker,
		retryAttempts: domain.SendEventDefaultRetryAttempts,
		timeout:       domain.SendEventDefaultTimeout,
	}
	for _, opt := range opts {
		if err := opt(ep); err != nil {
			return nil, err
		}
	}
	return ep, nil
}

func (ep *EventProducer) Fire(ctx context.Context, event any) {
	taskHandler := func(tctx context.Context) {
		retryHandler := func() error {
			key := uuid.NewString()
			data, err := json.Marshal(&event)
			if err != nil {
				return fmt.Errorf("'json.Marshal' failed: %w", err)
			}

			cbHandler := func() ([]byte, error) {
				logger := logging.Get(tctx)
				logger.Debug("sending event attempt started", "event", event)
				defer logger.Debug("sending event attempt ended", "event", event)

				dctx, cancel := context.WithTimeout(tctx, ep.timeout)
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
			retry.Context(tctx),
		); err != nil {
			logger := logging.Get(tctx)
			logger.Error("send event failed", "event", event, "error", err)
		}
	}
	ep.orch.Go(ctx, taskHandler)
}
