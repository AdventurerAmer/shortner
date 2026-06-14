package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AdventurerAmer/shortner/config"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/web"
	"github.com/google/uuid"
)

type Handlers struct {
	cfg           *config.ServiceConfig
	shorteningSrv ports.ShorteningService
}

func NewHandlers(shorteningSrv ports.ShorteningService) *Handlers {
	return &Handlers{
		shorteningSrv: shorteningSrv,
	}
}

func (h *Handlers) Shorten(c *web.Context) (any, error) {
	var req ports.ShortenURLRequest
	if err := c.Bind(&req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(c.Ctx(), h.cfg.DefaultTimeout)
	defer cancel()

	userId := uuid.NewString()
	resp, err := h.shorteningSrv.Shorten(ctx, userId, req)
	if err != nil {
		return nil, fmt.Errorf("failed to shorten url: %w", err)
	}

	c.SetStatus(http.StatusCreated)
	return resp, nil
}
