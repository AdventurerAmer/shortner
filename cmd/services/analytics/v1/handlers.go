package v1

import (
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/web"
)

type handlers struct {
	service ports.AnalyticService
}

func newHandlers(service ports.AnalyticService) *handlers {
	return &handlers{
		service: service,
	}
}

func (h *handlers) get(c *web.Context) (any, error) {
	req := ports.GetAnalyticRequest{
		Alias: c.Request.PathValue("alias"),
	}

	resp, err := h.service.Get(c.Ctx(), req)
	if err != nil {
		return nil, fmt.Errorf("'service.Get' failed: %w", err)
	}

	return resp, nil
}

func (h *handlers) clicks(c *web.Context) (any, error) {
	req := ports.IncrementClicksRequest{
		Alias:  c.Request.PathValue("alias"),
		Clicks: 1,
	}

	resp, err := h.service.IncrementClicks(c.Ctx(), req)
	if err != nil {
		return nil, fmt.Errorf("'service.IncrementClicks' failed: %w", err)
	}

	return resp, nil
}
