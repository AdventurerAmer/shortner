package users

import (
	"context"
	"fmt"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/validation"
)

type Config struct {
	UsersRepo ports.UsersRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.UsersService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Get(ctx context.Context, req ports.GetUserRequest) (ports.GetUserResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.GetUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.Get(ctx, req.Id)
	if err != nil {
		return ports.GetUserResponse{}, fmt.Errorf("'UsersRepo.Get' failed: %w", err)
	}

	resp := ports.GetUserResponse{
		User: domain.NewFrontendUser(user),
	}
	return resp, nil
}

func (srv *service) Update(ctx context.Context, req ports.UpdateUserRequest) (ports.UpdateUserResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.Get(ctx, req.Id)
	if err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("'UsersRepo.Get' failed: %w", err)
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	user.UpdatedAt = time.Now().UTC()

	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.UpdateUserResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	resp := ports.UpdateUserResponse{
		User: domain.NewFrontendUser(user),
	}
	return resp, nil
}

func (srv *service) Delete(ctx context.Context, req ports.DeleteUserRequest) (ports.DeleteUserResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.DeleteUserResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	if err := srv.UsersRepo.Delete(ctx, req.Id); err != nil {
		return ports.DeleteUserResponse{}, fmt.Errorf("'UsersRepo.Delete' failed: %w", err)
	}
	return ports.DeleteUserResponse{}, nil
}
