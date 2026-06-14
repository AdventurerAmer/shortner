package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type URLMappingRepository interface {
	Create(ctx context.Context, mapping *domain.URLMapping) error
	Get(ctx context.Context, shortURL string) (*domain.URLMapping, error)
	Delete(ctx context.Context, shortURL string) error
}

type ShorteningService interface {
	Shorten(ctx context.Context, userId string, req ShortenURLRequest) (ShortenURLResponse, error)
}

type ShortenURLRequest struct {
	LongURL string `json:"longURL"`
}

type ShortenURLResponse struct {
	CreatedAt time.Time `json:"createdAt"`
	ShortURL  string    `json:"shortURL"`
}

type RedirectingService interface {
	Redirect(ctx context.Context, req RedirectRequest) (RedirectResponse, error)
}

type RedirectRequest struct {
	ShortURL string `json:"shortURL"`
}

type RedirectResponse struct {
	LongURL string `json:"longURL"`
}
