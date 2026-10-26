package ports

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type UsersRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
