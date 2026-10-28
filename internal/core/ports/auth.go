package ports

import (
	"context"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (RegisterResponse, *domain.User, error)
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	VerifyToken(ctx context.Context, req VerifyTokenRequest) (VerifyTokenResponse, error)

	SendVerification(ctx context.Context, req SendVerificationRequest) (SendVerificationResponse, error)
	VerifyUser(ctx context.Context, req VerifyUserRequest) (VerifyResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) (ResetPasswordResponse, error)
	ChangePassword(ctx context.Context, req ChangePasswordRequest) (ChangePasswordResponse, error)
}

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"displayName" validate:"required,min=3,max=32"`
	Password    string `json:"password" validate:"required,len=8,max=128"`
}

type RegisterResponse struct {
	User *domain.FrontendUser `json:"user"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,len=8,max=128"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type VerifyTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type VerifyTokenResponse struct {
	UserId string `json:"userId"`
}

type VerifyUserRequest struct {
	Token string `json:"token" form:"token" validate:"required,len=43"`
}

type VerifyResponse struct {
	Message string `json:"message"`
}

type SendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SendVerificationResponse struct {
	Message string `json:"message"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required,len=43"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type ChangePasswordRequest struct {
	UserId          string `json:"userId" validate:"required"`
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,max=128"`
}

type ChangePasswordResponse struct {
	Message string `json:"message"`
}
