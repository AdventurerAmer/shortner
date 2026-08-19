package v1

import (
	"fmt"
	"net/http"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/telemetry"
	"github.com/google/uuid"
)

type handlers struct {
	service ports.ShorteningService
}

func NewHandlers(service ports.ShorteningService) *handlers {
	return &handlers{
		service: service,
	}
}

func (h *handlers) shorten(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: shorten")
	defer span.End()

	var req ports.ShortenURLRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, fmt.Errorf("'c.BindJSON' failed: %w", err)
	}

	userId := uuid.NewString() // @Temprary: using uuid for now...
	resp, err := h.service.Shorten(dctx, userId, req)
	if err != nil {
		return nil, fmt.Errorf("'service.Shorten' failed: %w", err)
	}

	c.SetStatus(http.StatusCreated)
	return resp, nil
}
