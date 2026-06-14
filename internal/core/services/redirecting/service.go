package redirecting

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
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
	mapping, err := srv.URLMappingRepo.Get(ctx, req.ShortURL)
	if err != nil {
		return ports.RedirectResponse{}, fmt.Errorf("'URLMappingRepo.Get' failed: %w", err)
	}
	resp := ports.RedirectResponse{
		LongURL: mapping.LongURL,
	}
	return resp, nil
}
