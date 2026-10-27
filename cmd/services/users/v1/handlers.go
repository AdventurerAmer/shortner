package v1

import (
	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
)

type handlers struct {
	service ports.UsersService
}

func newHandlers(service ports.UsersService) *handlers {
	return &handlers{
		service: service,
	}
}

func (h *handlers) register(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) get(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) update(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) delete(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) verify(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) sendVerification(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) forgotPassword(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) resetPassword(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}

func (h *handlers) changePassword(c *web.Context) (any, error) {
	// dctx, span := telemetry.NewSpan(c.Ctx(), "handler: redirect")
	// defer span.End()
	return nil, nil
}
