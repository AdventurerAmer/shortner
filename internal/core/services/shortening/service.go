package shortening

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	ShortURLPrefix string
	Shard          string
	URLMappingRepo ports.URLMappingRepository
	Snowflake      *domain.Snowflake
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
	if err := validation.Validate(&req); err != nil {
		return ports.ShortenURLResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	alias := srv.Snowflake.NextBase62(srv.Shard)
	mapping := &domain.URLMapping{
		Alias:     alias,
		LongURL:   req.LongURL,
		CreatedAt: time.Now().UTC(),
		UserId:    userId,
	}
	if err := srv.URLMappingRepo.Create(ctx, mapping); err != nil {
		return ports.ShortenURLResponse{}, fmt.Errorf("'URLMappingRep.Create' failed: %w", err)
	}
	shortURL := srv.ShortURLPrefix + alias
	resp := ports.ShortenURLResponse{
		CreatedAt: mapping.CreatedAt,
		ShortURL:  shortURL,
	}
	return resp, nil
}
