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
	logger         *logging.Logger
	cfg            *config.ServiceConfig
	redirectingSrv ports.RedirectingService
}

func NewHandlers(logger *logging.Logger, cfg *config.ServiceConfig, redirectingSrv ports.RedirectingService) *Handlers {
	return &Handlers{
		logger:         logger,
		cfg:            cfg,
		redirectingSrv: redirectingSrv,
	}
}

func (h *Handlers) Redirect(c *web.Context) (any, error) {
	shortURL := c.Request.PathValue("short_url")
	h.logger.Debug("Redirect Request", "shortURL", shortURL)
	req := ports.RedirectRequest{
		ShortURL: shortURL,
	}
	ctx, cancel := context.WithTimeout(c.Ctx(), h.cfg.DefaultTimeout)
	defer cancel()

	resp, err := h.redirectingSrv.Redirect(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to redirect url: %w", err)
	}

	return resp, nil
}
