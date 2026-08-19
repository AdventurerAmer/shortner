package redirecting

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/telemetry"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	URLMappingRepo ports.URLMappingRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.RedirectingService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Redirect(ctx context.Context, req ports.RedirectRequest) (ports.RedirectResponse, error) {
	dctx, span := telemetry.NewSpan(ctx, "Redirecting Service: Redirect")
	defer span.End()

	if err := validation.Validate(&req); err != nil {
		span.RecordError(err)
		return ports.RedirectResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	mapping, err := srv.URLMappingRepo.Get(dctx, req.Alias)
	if err != nil {
		span.RecordError(err)
		return ports.RedirectResponse{}, fmt.Errorf("'URLMappingRepo.Get' failed: %w", err)
	}
	resp := ports.RedirectResponse{
		LongURL: mapping.LongURL,
	}
	return resp, nil
}
