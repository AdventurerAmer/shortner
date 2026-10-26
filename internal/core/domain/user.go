package domain

import "time"

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
