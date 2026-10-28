package domain

import (
	"time"
)

type User struct {
	Id           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash"`
	DisplayName  string    `json:"displayName"`
	IsVerified   bool      `json:"isVerified"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Version      int       `json:"version"`
}

type FrontendUser struct {
	Email       string    `json:"email"`
	Displayname string    `json:"displayName"`
	IsVerified  bool      `json:"isVerified"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Version     int       `json:"version"`
}

func NewFrontendUser(user *User) *FrontendUser {
	return &FrontendUser{
		Email:       user.Email,
		Displayname: user.DisplayName,
		IsVerified:  user.IsVerified,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Version:     user.Version,
	}
}
