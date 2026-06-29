package v1

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"
)

type Handlers struct {
	logger *logging.Logger
	cfg    *config.ServiceConfig
	srv    ports.AnalyticService
}

func NewHandlers(logger *logging.Logger, cfg *config.ServiceConfig, srv ports.AnalyticService) *Handlers {
	return &Handlers{
		logger: logger,
		cfg:    cfg,
		srv:    srv,
	}
}

func (h *Handlers) Clicks(c *web.Context) (any, error) {
	req := ports.IncrementClicksRequest{
		Alias:  c.Request.PathValue("alias"),
		Clicks: 1,
	}

	ctx, cancel := context.WithTimeout(c.Ctx(), h.cfg.DefaultTimeout)
	defer cancel()

	resp, err := h.srv.IncrementClicks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("'srv.IncrementClicks' failed: %w", err)
	}

	return resp, nil
}
