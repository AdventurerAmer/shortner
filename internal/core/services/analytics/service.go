package analytics

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/telemetry"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	AnalyticClicksRepo ports.AnalyticsRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.AnalyticsService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Get(ctx context.Context, req ports.GetAnalyticRequest) (ports.GetAnalyticResponse, error) {
	dctx, span := telemetry.NewSpan(ctx, "Analytics Service: Get")
	defer span.End()

	if err := validation.Validate(&req); err != nil {
		span.RecordError(err)
		return ports.GetAnalyticResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	stat, err := srv.AnalyticClicksRepo.Get(dctx, req.Alias)
	if err != nil {
		span.RecordError(err)
		return ports.GetAnalyticResponse{}, fmt.Errorf("'AnalyticStatRepo.Get' failed: %w", err)
	}
	resp := ports.GetAnalyticResponse{
		Analytic: stat,
	}
	return resp, nil
}
