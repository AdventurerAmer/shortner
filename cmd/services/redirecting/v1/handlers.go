package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/telemetry"
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

func (h *handlers) redirect(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	defer span.End()

	req := ports.RedirectRequest{
		Alias: c.Request.PathValue("alias"),
	}

	resp, err := h.service.Redirect(dctx, req)
	if err != nil {
		return nil, fmt.Errorf("'service.Redirect' failed: %w", err)
	}

	event := domain.ClickEvent{
		Alias:     req.Alias,
		Timestamp: time.Now().UTC(),
	}
	h.eventProducer.Fire(dctx, event)

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)
	return nil, nil
}
