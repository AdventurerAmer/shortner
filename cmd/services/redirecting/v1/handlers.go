package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/async"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/web"

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

	event := domain.ClickEvent{
		Alias:     req.Alias,
		Timestamp: time.Now().UTC(),
	}

	eventProducer := ports.DefaultEventProducer(h.producer)

	goFunc := func(ctx context.Context) {
		eventProducer.Fire(ctx, event)
	}
	h.orch.Go(c.Ctx(), goFunc)

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)
	return nil, nil
}
