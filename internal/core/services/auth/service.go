package auth

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
	"github.com/golang-jwt/jwt/v5"
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

	JWTSecret            string
	JWTIssuer            string
	JWTTokenExpiresAfter time.Duration

	UsersRepo ports.UsersRepository
}

type service struct {
	Config
}

func New(cfg Config) ports.AuthService {
	return &service{Config: cfg}
}

func (srv *service) Register(ctx context.Context, req ports.RegisterRequest) (ports.RegisterResponse, *domain.User, error) {
	if err := validation.Validate(req); err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("validation failed: %w", err)
	}

	passwordHash, err := srv.PasswordManager.Hash(req.Password)
	if err != nil {
		return ports.RegisterResponse{}, nil, fmt.Errorf("'PasswordManager.Hash' failed: %w", err)
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
		return ports.RegisterResponse{}, nil, fmt.Errorf("transaction failed: %w", err)
	}

	resp := ports.RegisterResponse{
		User: domain.NewFrontendUser(user),
	}
	return resp, user, nil
}

func (srv *service) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.LoginResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	email := strings.ToLower(req.Email)
	user, err := srv.UsersRepo.GetByEmail(ctx, email)
	if err != nil {
		if errs.IsNotFound(err) {
			return ports.LoginResponse{}, fmt.Errorf("invalid credentials")
		}
		return ports.LoginResponse{}, fmt.Errorf("'UsersRepo.GetByEmail' failed: %w", err)
	}

	ok, err := srv.PasswordManager.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return ports.LoginResponse{}, fmt.Errorf("'PasswordManager.Verify' failed: %w", err)
	}

	if !ok {
		return ports.LoginResponse{}, fmt.Errorf("invalid credentials")
	}

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   user.Id,
		Issuer:    srv.JWTIssuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(srv.JWTTokenExpiresAfter)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(srv.JWTSecret)
	if err != nil {
		return ports.LoginResponse{}, fmt.Errorf("'token.SignedString' failed: %w", err)
	}
	resp := ports.LoginResponse{
		Token: tokenStr,
	}
	return resp, nil
}

func (srv *service) VerifyToken(ctx context.Context, req ports.VerifyTokenRequest) (ports.VerifyTokenResponse, error) {
	if err := validation.Validate(req); err != nil {
		return ports.VerifyTokenResponse{}, fmt.Errorf("validation failed: %w", err)
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return srv.JWTSecret, nil
	}

	token, err := jwt.ParseWithClaims(req.Token, &jwt.RegisteredClaims{}, keyFunc)
	if err != nil {
		return ports.VerifyTokenResponse{}, fmt.Errorf("'jwt.ParseWithClaims' failed: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return ports.VerifyTokenResponse{}, fmt.Errorf("invalid or expired token")
	}
	resp := ports.VerifyTokenResponse{
		UserId: claims.Subject,
	}
	return resp, nil
}

func (srv *service) VerifyUser(ctx context.Context, req ports.VerifyUserRequest) (ports.VerifyResponse, error) {
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
