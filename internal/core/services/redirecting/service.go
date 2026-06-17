package redirecting

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/logging"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	URLMappingRepo ports.URLMappingRepository
}

type service struct {
	Config
	logger *logging.Logger
}

func New(logger *logging.Logger, cfg Config) ports.RedirectingService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Redirect(ctx context.Context, req ports.RedirectRequest) (ports.RedirectResponse, error) {
	if err := validation.Validate(&req); err != nil {
		return ports.RedirectResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	mapping, err := srv.URLMappingRepo.Get(ctx, req.Alias)
	if err != nil {
		return ports.RedirectResponse{}, fmt.Errorf("'URLMappingRepo.Get' failed: %w", err)
	}
	resp := ports.RedirectResponse{
		LongURL: mapping.LongURL,
	}
	return resp, nil
}
