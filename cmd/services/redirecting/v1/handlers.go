package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"

	analyticsV1 "github.com/AdventurerAmer/shortner/cmd/services/analytics/v1"
)

type Handlers struct {
	cfg             *config.ServiceConfig
	srv             ports.RedirectingService
	analyticsClient *analyticsV1.Client
}

func NewHandlers(cfg *config.ServiceConfig, srv ports.RedirectingService, analyticsClient *analyticsV1.Client) *Handlers {
	return &Handlers{
		cfg:             cfg,
		srv:             srv,
		analyticsClient: analyticsClient,
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
		// if errs.IsNotFound(err) {
		// } else {
		// }
		return nil, fmt.Errorf("'srv.Redirect' failed: %w", err)
	}

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)

	go func() {
		alias := req.Alias
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := h.analyticsClient.IncrementClicks(dctx, alias); err != nil {
			logger := logging.Get(ctx)
			logger.Error("failed to increment clicks", "error", err)
		}
	}()

	return nil, nil
}
