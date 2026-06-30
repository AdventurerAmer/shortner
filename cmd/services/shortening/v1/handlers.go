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
	cfg *config.ServiceConfig
	srv ports.ShorteningService
}

func NewHandlers(cfg *config.ServiceConfig, srv ports.ShorteningService) *Handlers {
	return &Handlers{
		cfg: cfg,
		srv: srv,
	}
}

func (h *Handlers) Shorten(c *web.Context) (any, error) {
	var req ports.ShortenURLRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(c.Ctx(), h.cfg.DefaultTimeout)
	defer cancel()

	userId := uuid.NewString() // @Temprary: using uuid for now...
	resp, err := h.srv.Shorten(ctx, userId, req)
	if err != nil {
		return nil, fmt.Errorf("'srv.Shorten' failed: %w", err)
	}

	c.SetStatus(http.StatusCreated)
	return resp, nil
}
