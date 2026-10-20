package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type AnalyticsRepository interface {
	Put(ctx context.Context, eventId, alias string, createdAt time.Time) error
	Get(ctx context.Context, alias string) (*domain.Analytic, error)
	Delete(ctx context.Context, eventId string) error
}

type AnalyticService interface {
	Get(ctx context.Context, req GetAnalyticRequest) (GetAnalyticResponse, error)
}

type GetAnalyticRequest struct {
	Alias string `json:"alias" validate:"required,len=9"`
}

type GetAnalyticResponse struct {
	Analytic *domain.Analytic `json:"analyticStat"`
}
