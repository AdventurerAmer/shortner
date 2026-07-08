package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdventurerAmer/shortner/async"
	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/web"

	analyticsV1 "github.com/AdventurerAmer/shortner/cmd/services/analytics/v1"

	"github.com/avast/retry-go"
	"github.com/sony/gobreaker/v2"
)

type handlers struct {
	service         ports.RedirectingService
	analyticsClient *analyticsV1.Client
	orch            *async.Orchestrator
}

func newHandlers(service ports.RedirectingService, analyticsClient *analyticsV1.Client, orch *async.Orchestrator) *handlers {
	return &handlers{
		service:         service,
		analyticsClient: analyticsClient,
		orch:            orch,
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
		alias := req.Alias

		retryFunc := func() error {

			_, err := analyticsCB.Execute(func() ([]byte, error) {
				// TODO: harcoding timeout here...
				dctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				if err := h.analyticsClient.IncrementClicks(dctx, alias); err != nil {
					if errs.IsNotFound(err) || errs.IsValidation(err) {
						return nil, nil
					}
					return nil, err
				}

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
			logger.Error("increment clicks failed", "alias", alias, "error", err)
		}
	}
	h.orch.Go(c.Ctx(), goFunc)

	// http.StatusFound represents a temporary (302) redirect
	http.Redirect(c.ResponseWriter, c.Request, resp.LongURL, http.StatusFound)
	return nil, nil
}
