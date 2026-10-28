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

type UsersService interface {
	Get(ctx context.Context, req GetUserRequest) (GetUserResponse, error)
	Update(ctx context.Context, req UpdateUserRequest) (UpdateUserResponse, error)
	Delete(ctx context.Context, req DeleteUserRequest) (DeleteUserResponse, error)
}

type GetUserRequest struct {
	Id string `json:"id" validate:"required"`
}

type GetUserResponse struct {
	User *domain.FrontendUser `json:"user"`
}

type UpdateUserRequest struct {
	Id          string  `json:"id" validate:"required"`
	DisplayName *string `json:"displayName" validate:"omitempty,min=3,max=32"`
}

type UpdateUserResponse struct {
	User *domain.FrontendUser `json:"user"`
}

type DeleteUserRequest struct {
	Id string `json:"id" validate:"required"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
}
