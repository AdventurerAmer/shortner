package ports

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type AnalyticStatRepository interface {
	Get(ctx context.Context, alias string) (*domain.AnalyticStat, error)
	Put(ctx context.Context, id string, aliases []string, clicks []int) error
	Delete(ctx context.Context, alias string) error
}

type AnalyticService interface {
	Get(ctx context.Context, req GetAnalyticStatRequest) (GetAnalyticStatResponse, error)
}

type GetAnalyticStatRequest struct {
	Alias string `json:"alias" validate:"required,len=9"`
}

type GetAnalyticStatResponse struct {
	AnalyticStat *domain.AnalyticStat `json:"analyticStat"`
}
