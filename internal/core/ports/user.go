package ports

import (
	"context"
	"time"

	"github.com/AdventurerAmer/shortner/internal/core/domain"
)

type FrontendUser struct {
	Email       string    `json:"email"`
	Displayname string    `json:"displayName"`
	IsVerified  bool      `json:"isVerified"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"version"`
}

func NewFrontendUser(user *domain.User) *FrontendUser {
	return &FrontendUser{
		Email:       user.Email,
		Displayname: user.DisplayName,
		IsVerified:  user.IsVerified,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Version:     user.Version,
	}
}

type UsersRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

type UsersService interface {
	Register(ctx context.Context, req RegisterUserRequest) (RegisterUserResponse, *domain.User, error)
	Get(ctx context.Context, req GetUserRequest) (GetUserResponse, error)
	Update(ctx context.Context, req UpdateUserRequest) (UpdateUserResponse, error)
	Delete(ctx context.Context, req DeleteUserRequest) (DeleteUserResponse, error)
	SendVerification(ctx context.Context, req SendVerificationRequest) (SendVerificationResponse, error)
	Verify(ctx context.Context, req VerifyUserRequest) (VerifyResponse, error)
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) (ResetPasswordResponse, error)
	ChangePassword(ctx context.Context, req ChangePasswordRequest) (ChangePasswordResponse, error)
}

type RegisterUserRequest struct {
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"displayName" validate:"required,min=3,max=32"`
	Password    string `json:"password" validate:"required,len=8,max=128"`
}

type RegisterUserResponse struct {
	User *FrontendUser `json:"user"`
}

type GetUserRequest struct {
	Id string `json:"id" validate:"required"`
}

type GetUserResponse struct {
	User *FrontendUser `json:"user"`
}

type UpdateUserRequest struct {
	Id          string  `json:"id" validate:"required"`
	DisplayName *string `json:"displayName" validate:"omitempty,min=3,max=32"`
}

type UpdateUserResponse struct {
	User *FrontendUser `json:"user"`
}

type DeleteUserRequest struct {
	Id string `json:"id" validate:"required"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
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
