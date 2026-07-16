package analytics

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	AnalyticStatRepo ports.AnalyticStatRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.AnalyticService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Get(ctx context.Context, req ports.GetAnalyticStatRequest) (ports.GetAnalyticStatResponse, error) {
	if err := validation.Validate(&req); err != nil {
		return ports.GetAnalyticStatResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	stat, err := srv.AnalyticStatRepo.Get(ctx, req.Alias)
	if err != nil {
		if errs.IsNotFound(err) {
			return ports.GetAnalyticStatResponse{}, err
		}
		return ports.GetAnalyticStatResponse{}, fmt.Errorf("'AnalyticStatRepo.Get' failed: %w", err)
	}
	resp := ports.GetAnalyticStatResponse{
		AnalyticStat: stat,
	}
	return resp, nil
}
