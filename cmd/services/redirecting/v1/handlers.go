package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"

	analyticsV1 "github.com/AdventurerAmer/shortner/cmd/services/analytics/v1"
)

type handlers struct {
	service         ports.RedirectingService
	analyticsClient *analyticsV1.Client
}

func newHandlers(service ports.RedirectingService, analyticsClient *analyticsV1.Client) *handlers {
	return &handlers{
		service:         service,
		analyticsClient: analyticsClient,
	}
}

func (h *handlers) redirect(c *web.Context) (any, error) {
	req := ports.RedirectRequest{
		Alias: c.Request.PathValue("alias"),
	}

	resp, err := h.service.Redirect(c.Ctx(), req)
	if err != nil {
		// TODO: html templates for not-found and err
		// if errs.IsNotFound(err) {
		// } else {
		// }
		return nil, fmt.Errorf("'service.Redirect' failed: %w", err)
	}

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)

	go func() {
		alias := req.Alias
		dctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := h.analyticsClient.IncrementClicks(dctx, alias); err != nil {
			logger := logging.Get(dctx)
			logger.Error("failed to increment clicks", "error", err)
		}
	}()

	return nil, nil
}
