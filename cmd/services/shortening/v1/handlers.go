package v1

import (
	"fmt"
	"net/http"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/web"
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
	var req ports.ShortenURLRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, fmt.Errorf("'c.BindJSON' failed: %w", err)
	}

	userId := uuid.NewString() // @Temprary: using uuid for now...
	resp, err := h.service.Shorten(c.Ctx(), userId, req)
	if err != nil {
		return nil, fmt.Errorf("'service.Shorten' failed: %w", err)
	}

	c.SetStatus(http.StatusCreated)
	return resp, nil
}
