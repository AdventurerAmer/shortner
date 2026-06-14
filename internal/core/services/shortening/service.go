package shortening

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/google/uuid"
)

type Config struct {
	ShortURLPrefix string
	URLMappingRepo ports.URLMappingRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.ShorteningService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Shorten(ctx context.Context, userId string, req ports.ShortenURLRequest) (ports.ShortenURLResponse, error) {
	// TODO: check if userId is valid
	// TODO: check if long url is valid
	shortURL := uuid.NewString() // USING uuid for now...
	mapping := &domain.URLMapping{
		ShortURL:  shortURL,
		LongURL:   req.LongURL,
		CreatedAt: time.Now().UTC(),
		UserId:    userId,
	}
	if err := srv.URLMappingRepo.Create(ctx, mapping); err != nil {
		return ports.ShortenURLResponse{}, fmt.Errorf("'URLMappingRep.Create' failed: %w", err)
	}
	resp := ports.ShortenURLResponse{
		CreatedAt: mapping.CreatedAt,
		ShortURL:  srv.ShortURLPrefix + shortURL,
	}
	return resp, nil
}
