package ports

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type AnalyicRepository interface {
	Create(ctx context.Context, analytic *domain.Analytic) error
	Get(ctx context.Context, alias string) (*domain.Analytic, error)
	Update(ctx context.Context, analytic *domain.Analytic) error
	Delete(ctx context.Context, alias string) error
}

type AnalyicService interface {
	Get(ctx context.Context)
	IncrementClicks(ctx context.Context, req IncrementClicksRequest) (IncrementClicksResponse, error)
}

type GetAnalyicRequest struct {
	Alias string `json:"alias"`
}

type GetAnalyicResponse struct {
	Analytic *domain.Analytic `json:"analytic"`
}

type IncrementClicksRequest struct {
}

type IncrementClicksResponse struct {
}
