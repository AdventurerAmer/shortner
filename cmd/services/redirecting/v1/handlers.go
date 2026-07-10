package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/async"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"

	"github.com/avast/retry-go"
	"github.com/sony/gobreaker/v2"
)

type handlers struct {
	service  ports.RedirectingService
	producer ports.Producer
	orch     *async.Orchestrator
}

func newHandlers(service ports.RedirectingService, producer ports.Producer, orch *async.Orchestrator) *handlers {
	return &handlers{
		service:  service,
		producer: producer,
		orch:     orch,
	}
}

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

func (h *handlers) redirect(c *web.Context) (any, error) {
	req := ports.RedirectRequest{
		Alias: c.Request.PathValue("alias"),
	}

	resp, err := h.service.Redirect(c.Ctx(), req)
	if err != nil {
		return nil, fmt.Errorf("'service.Redirect' failed: %w", err)
	}

	goFunc := func(ctx context.Context) {
		event := domain.ClickEvent{
			Alias:     req.Alias,
			Timestamp: time.Now().UTC(),
		}

		retryFunc := func() error {

			_, err := analyticsCB.Execute(func() ([]byte, error) {
				dctx, cancel := context.WithTimeout(ctx, domain.ClickEventTimeout)
				defer cancel()

				key := event.Alias
				data, err := json.Marshal(&event)
				if err != nil {
					return nil, fmt.Errorf("'json.Marshal' failed: %w", err)
				}

				if err := h.producer.Send(dctx, key, data); err != nil {
					return nil, fmt.Errorf("'producer.Send' failed: %w", err)
				}

				logger := logging.Get(ctx)
				logger.Debug("sending clicks event succeeded", "event", event)

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
	}
	h.orch.Go(c.Ctx(), goFunc)

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)
	return nil, nil
}
