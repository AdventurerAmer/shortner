package ports

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type AnalyticClicksRepository interface {
	Get(ctx context.Context, alias string) (*domain.AnalyticClicks, error)
	Put(ctx context.Context, ids []string, aliases []string, clicks []int) error
	Delete(ctx context.Context, alias string) error
}

type AnalyticService interface {
	Get(ctx context.Context, req GetAnalyticStatRequest) (GetAnalyticStatResponse, error)
}

type GetAnalyticStatRequest struct {
	Alias string `json:"alias" validate:"required,len=9"`
}

type GetAnalyticStatResponse struct {
	AnalyticStat *domain.AnalyticClicks `json:"analyticStat"`
}
