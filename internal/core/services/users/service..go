package users

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/tokens"
	"github.com/AdventurerAmer/shortner/validation"
	"github.com/google/uuid"
)

type Config struct {
	PasswordManager          ports.PasswordManager
	Producer                 ports.Producer
	Transactor               ports.Transactor
	VerificationTokenManager *tokens.Manager
	UsersRepo                ports.UsersRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.UsersService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Register(ctx context.Context, req ports.RegisterUserRequest) (ports.RegisterUserResponse, *domain.User, error) {
	if err := validation.Validate(req); err != nil {
		return ports.RegisterUserResponse{}, nil, fmt.Errorf("validation failed: %w", err)
	}

	passwordHash, err := srv.PasswordManager.Hash(req.Password)
	if err != nil {
		return ports.RegisterUserResponse{}, nil, fmt.Errorf("'PasswordManager.Hash' failed: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		DisplayName:  strings.TrimSpace(req.DisplayName),
		PasswordHash: passwordHash,
		IsVerified:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	txn := func(tctx context.Context) error {
		if err := srv.UsersRepo.Create(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Create' failed: %w", err)
		}

		token, plain, err := srv.VerificationTokenManager.Create(tctx, user.Id)
		if err != nil {
			return fmt.Errorf("'VerificationTokenManager.Create' failed: %w", err)
		}

		eventId := uuid.NewString()
		event := domain.UserCreatedEvent{
			Id:                         eventId,
			UserId:                     user.Id,
			VerificationToken:          plain,
			VerificationTokenExpiresAt: token.ExpiresAt,
		}
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("'json.Marshal' failed: %w", err)
		}
		if err := srv.Producer.Send(tctx, eventId, data); err != nil {
			return fmt.Errorf("'EventPublisher.Publish' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.RegisterUserResponse{}, nil, fmt.Errorf("transaction failed: %w", err)
	}

	resp := ports.RegisterUserResponse{
		User: ports.NewFrontendUser(user),
	}
	return resp, user, nil
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
		User: ports.NewFrontendUser(user),
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
		User: ports.NewFrontendUser(user),
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
