package v1

import (
	"fmt"

	"github.com/AdventurerAmer/shortner/apps/web"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/telemetry"
)

type handlers struct {
	service ports.AuthService
}

func newHandlers(service ports.AuthService) *handlers {
	return &handlers{
		service: service,
	}
}

func (h *handlers) register(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: register")
	defer span.End()

	var req ports.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}

	resp, _, err := h.service.Register(dctx, req)
	if err != nil {
		return nil, fmt.Errorf("'service.Register' failed: %w", err)
	}

	return resp, nil
}

func (h *handlers) login(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: login")
	defer span.End()

	var req ports.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}

	resp, err := h.service.Login(dctx, req)
	if err != nil {
		return nil, fmt.Errorf("'service.Login' failed: %w", err)
	}

	return resp, nil
}

func (h *handlers) sendVerification(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: sendVerification")
	defer span.End()

	var req ports.SendVerificationRequest
	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}

	resp, err := h.service.SendVerification(dctx, req)
	if err != nil {
		return nil, fmt.Errorf("'service.SendVerification' failed: %w", err)
	}

	return resp, nil
}

func (h *handlers) verify(c *web.Context) (any, error) {
	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: verify")
	defer span.End()

	token := c.Request.URL.Query().Get("token")
	req := ports.VerifyUserRequest{Token: token}
	resp, err := h.service.VerifyUser(dctx, req)
	if err != nil {
		return nil, fmt.Errorf("'service.VerifyUser' failed: %w", err)
	}

	return resp, nil
}

// func (h *handlers) forgotPassword(c *web.Context) (any, error) {
// 	dctx, span := telemetry.NewSpan(c.Ctx(), "handler: forgotPassword")
// 	defer span.End()

// 	var req ports.ForgotPasswordRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		return nil, err
// 	}

// 	resp, err := h.service.ForgotPassword(dctx, req)
// 	if err != nil {
// 		return nil, fmt.Errorf("'service.ForgotPassword' failed: %w", err)
// 	}

// 	return resp, nil
// }
