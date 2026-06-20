package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"
)

type Handlers struct {
	logger *logging.Logger
	cfg    *config.ServiceConfig
	srv    ports.RedirectingService
}

func NewHandlers(logger *logging.Logger, cfg *config.ServiceConfig, srv ports.RedirectingService) *Handlers {
	return &Handlers{
		logger: logger,
		cfg:    cfg,
		srv:    srv,
	}
}

func (h *Handlers) Redirect(c *web.Context) (any, error) {
	req := ports.RedirectRequest{
		Alias: c.Request.PathValue("alias"),
	}

	ctx, cancel := context.WithTimeout(c.Ctx(), h.cfg.DefaultTimeout)
	defer cancel()

	resp, err := h.srv.Redirect(ctx, req)
	if err != nil {
		// TODO: html templates for not-found and err
		if errs.IsNotFound(err) {
		} else {
		}
		return nil, fmt.Errorf("'srv.Redirect' failed: %w", err)
	}

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)

	return nil, nil
}
