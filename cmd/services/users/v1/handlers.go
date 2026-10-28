package v1

import (
	"fmt"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/telemetry"
)

type handlers struct {
	service ports.UsersService
}

func newHandlers(service ports.UsersService) *handlers {
	return &handlers{
		service: service,
	}
}

func (h *handlers) get(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: get")
	defer span.End()

	userId := userIdFromContext(dctx)
	if userId == "" {
		return nil, fmt.Errorf("user is not authenticated")
	}

	resp, err := h.service.Get(dctx, ports.GetUserRequest{Id: userId})
	if err != nil {
		return nil, fmt.Errorf("'service.Get' failed: %w", err)
	}

	return resp, nil
}

func (h *handlers) update(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: update")
	defer span.End()

	userId := userIdFromContext(dctx)
	if userId == "" {
		return nil, fmt.Errorf("user is not authenticated")
	}

	var req ports.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}
	resp, err := h.service.Update(dctx, req)
	if err != nil {
		return nil, fmt.Errorf("'service.Update' failed: %w", err)
	}

	return resp, nil
}
