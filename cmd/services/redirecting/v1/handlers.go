package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/web"

	"github.com/sony/gobreaker/v2"
)

type handlers struct {
	service       ports.RedirectingService
	eventProducer *ports.EventProducer
}

func newHandlers(service ports.RedirectingService, eventProducer *ports.EventProducer) *handlers {
	return &handlers{
		service:       service,
		eventProducer: eventProducer,
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
	h.eventProducer.Fire(c.Ctx(), event)

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)
	return nil, nil
}
