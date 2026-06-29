package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	AnalyticRepo ports.AnalyticRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.AnalyticService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Get(ctx context.Context, req ports.GetAnalyticRequest) (ports.GetAnalyticResponse, error) {
	if err := validation.Validate(&req); err != nil {
		return ports.GetAnalyticResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	a, err := srv.AnalyticRepo.Get(ctx, req.Alias)
	if err != nil {
		if errs.IsNotFound(err) {
			return ports.GetAnalyticResponse{}, err
		}
		return ports.GetAnalyticResponse{}, fmt.Errorf("'AnalyticRepo.Get' failed: %w", err)
	}
	resp := ports.GetAnalyticResponse{
		Analytic: a,
	}
	return resp, nil
}

func (srv *service) IncrementClicks(ctx context.Context, req ports.IncrementClicksRequest) (ports.IncrementClicksResponse, error) {
	if err := validation.Validate(&req); err != nil {
		return ports.IncrementClicksResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	a, err := srv.AnalyticRepo.Get(ctx, req.Alias)
	if err != nil {
		if !errs.IsNotFound(err) {
			return ports.IncrementClicksResponse{}, fmt.Errorf("'AnalyticRepo.Get' failed: %w", err)
		}
	}

	now := time.Now().UTC()

	if a == nil {
		a = &domain.Analytic{
			Alias:     req.Alias,
			CreatedAt: now,
			UpdatedAt: now,
			Clicks:    req.Clicks,
		}
		if err := srv.AnalyticRepo.Create(ctx, a); err != nil {
			return ports.IncrementClicksResponse{}, fmt.Errorf("'AnalyticRepo.Create' failed: %w", err)
		}
	} else {
		a.Clicks += req.Clicks
		a.UpdatedAt = now
		if err := srv.AnalyticRepo.Update(ctx, a); err != nil {
			return ports.IncrementClicksResponse{}, fmt.Errorf("'AnalyticRepo.Update' failed: %w", err)
		}
	}

	resp := ports.IncrementClicksResponse{
		Analytic: a,
	}
	return resp, nil
}
