package ports

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type AnalyticRepository interface {
	Create(ctx context.Context, analytic *domain.Analytic) error
	Get(ctx context.Context, alias string) (*domain.Analytic, error)
	Update(ctx context.Context, analytic *domain.Analytic) error
	Delete(ctx context.Context, alias string) error
}

type AnalyticService interface {
	Get(ctx context.Context, req GetAnalyticRequest) (GetAnalyticResponse, error)
	IncrementClicks(ctx context.Context, req IncrementClicksRequest) (IncrementClicksResponse, error)
}

type GetAnalyticRequest struct {
	Alias string `json:"alias" validate:"required,len=9"`
}

type GetAnalyticResponse struct {
	Analytic *domain.Analytic `json:"analytic"`
}

type IncrementClicksRequest struct {
	Alias  string `json:"alias" validate:"required,len=9"`
	Clicks int    `json:"clicks" validate:"required,min=1"`
}

type IncrementClicksResponse struct {
	Analytic *domain.Analytic `json:"analytic"`
}
