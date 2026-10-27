package users

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AdventurerAmer/shortner/errs"
	"github.com/AdventurerAmer/shortner/internal/core/domain"
	"github.com/AdventurerAmer/shortner/internal/core/ports"
	"github.com/AdventurerAmer/shortner/internal/tokens"
	"github.com/AdventurerAmer/shortner/validation"
	"github.com/google/uuid"
)

type Config struct {
	PasswordManager           ports.PasswordManager
	UserCreatedProducer       ports.Producer
	UserVerficiationProducer  ports.Producer
	UserPasswordResetProducer ports.Producer
	Transactor                ports.Transactor
	VerificationTokenManager  *tokens.Manager
	PasswordResetTokenManager *tokens.Manager
	UsersRepo                 ports.UsersRepository
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
		if err := srv.UserCreatedProducer.Send(tctx, eventId, data); err != nil {
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

func (srv *service) Verify(ctx context.Context, req ports.VerifyUserRequest) (ports.VerifyResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	ok, token, err := srv.VerificationTokenManager.Check(ctx, req.Token)
	if err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("VerificationTokenManager.Check")
	}

	if !ok {
		return ports.VerifyResponse{}, errs.NewFailedPrecondition("invalid or expired token")
	}

	txn := func(tctx context.Context) error {
		user, err := srv.UsersRepo.Get(tctx, token.UserId)
		if err != nil {
			return fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
		}

		now := time.Now().UTC()
		user.IsVerified = true
		user.UpdatedAt = now
		if err := srv.UsersRepo.Update(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		if err := srv.VerificationTokenManager.Invalidate(ctx, token); err != nil {
			return fmt.Errorf("'VerificationTokenManager.Update' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.VerifyResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.VerifyResponse{
		Message: "User was verified successfully",
	}, nil
}

func (srv *service) SendVerification(ctx context.Context, req ports.SendVerificationRequest) (ports.SendVerificationResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	if user.IsVerified {
		return ports.SendVerificationResponse{}, errs.NewFailedPrecondition("user is already verified")
	}

	ok, err := srv.VerificationTokenManager.IsLastTokenStillValid(ctx, user.Id)
	if err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("'VerificationTokenManager.IsLastTokenStillValid' failed: %w", err)
	}

	if ok {
		return ports.SendVerificationResponse{}, errs.NewFailedPrecondition("email was already sent")
	}

	txn := func(tctx context.Context) error {
		_, plain, err := srv.VerificationTokenManager.Create(ctx, user.Id)
		if err != nil {
			return fmt.Errorf("'VerificationTokenManager.Create' failed: %w", err)
		}

		e := domain.UserVerificationEvent{
			Id:     uuid.NewString(),
			UserId: user.Id,
			Token:  plain,
		}
		data, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("'json.Marshal' failed: %w", err)
		}
		if err := srv.UserVerficiationProducer.Send(ctx, e.Id, data); err != nil {
			return fmt.Errorf("'UserVerficiationProducer.Send' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.SendVerificationResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.SendVerificationResponse{
		Message: "Email was sent successfully",
	}, nil
}

func (srv *service) ForgotPassword(ctx context.Context, req ports.ForgotPasswordRequest) (ports.ForgotPasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	ok, err := srv.PasswordResetTokenManager.IsLastTokenStillValid(ctx, user.Id)
	if err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("'PasswordResetTokenManager.IsLastTokenStillValid' failed: %w", err)
	}

	if ok {
		return ports.ForgotPasswordResponse{}, errs.NewFailedPrecondition("email was already sent")
	}

	txn := func(tctx context.Context) error {
		_, plain, err := srv.PasswordResetTokenManager.Create(ctx, user.Id)
		if err != nil {
			return fmt.Errorf("'PasswordResetTokenManager.Create' failed: %w", err)
		}

		e := domain.UserPasswordResetEvent{
			Id:     uuid.NewString(),
			UserId: user.Id,
			Token:  plain,
		}
		data, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("'json.Marshal' failed: %w", err)
		}
		if err := srv.UserPasswordResetProducer.Send(ctx, e.Id, data); err != nil {
			return fmt.Errorf("'UserVerficiationProducer.Send' failed: %w", err)
		}

		return nil
	}
	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.ForgotPasswordResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.ForgotPasswordResponse{
		Message: "email was send successfully",
	}, nil
}

func (srv *service) ResetPassword(ctx context.Context, req ports.ResetPasswordRequest) (ports.ResetPasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	ok, token, err := srv.PasswordResetTokenManager.Check(ctx, req.Token)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'PasswordResetTokenManager.Check' failed: %w", err)
	}

	if !ok {
		return ports.ResetPasswordResponse{}, errs.NewFailedPrecondition("invalid or expired token")
	}

	passwordHash, err := srv.PasswordManager.Hash(req.Password)
	if err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("'PasswordHasher.Hash' failed: %w", err)
	}

	txn := func(tctx context.Context) error {
		user, err := srv.UsersRepo.Get(tctx, token.UserId)
		if err != nil {
			return fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
		}

		now := time.Now().UTC()
		user.PasswordHash = passwordHash
		user.UpdatedAt = now
		if err := srv.UsersRepo.Update(tctx, user); err != nil {
			return fmt.Errorf("'UsersRepo.Update' failed: %w", err)
		}

		if err := srv.PasswordResetTokenManager.Invalidate(ctx, token); err != nil {
			return fmt.Errorf("'PasswordResetTokenManager.Invalidate' failed: %w", err)
		}

		return nil
	}

	if err := srv.Transactor.WithTransaction(ctx, txn); err != nil {
		return ports.ResetPasswordResponse{}, fmt.Errorf("transaction failed: %w", err)
	}

	return ports.ResetPasswordResponse{
		Message: "Password was successfully reset",
	}, nil
}

func (srv *service) ChangePassword(ctx context.Context, req ports.ChangePasswordRequest) (ports.ChangePasswordResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	user, err := srv.UsersRepo.Get(ctx, req.UserId)
	if err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'UsersRepo.GetById' failed: %w", err)
	}

	ok, err := srv.PasswordManager.Verify(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'PasswordManager.Verify' failed: %w", err)
	}
	if !ok {
		return ports.ChangePasswordResponse{}, errs.NewFailedPrecondition("current password isn't correct")
	}

	hash, err := srv.PasswordManager.Hash(req.NewPassword)
	if err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'PasswordManager.Hash' failed: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now().UTC()
	if err := srv.UsersRepo.Update(ctx, user); err != nil {
		return ports.ChangePasswordResponse{}, fmt.Errorf("'UsersRepo.Update' failed: %w", err)
	}

	return ports.ChangePasswordResponse{
		Message: "Password was changed successfully",
	}, nil
}
