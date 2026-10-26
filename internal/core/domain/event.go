package domain

import (
	"time"
)

type Topic string

const (
	ClicksTopic Topic = "clicks"
	UsersTopic  Topic = "users"
)

type ClickEvent struct {
	Id        string    `json:"id"`
	Alias     string    `json:"alias"`
	Timestamp time.Time `json:"timestamp"`
}

type UserCreatedEvent struct {
	Id                         string    `json:"id"`
	UserId                     string    `json:"userId"`
	VerificationToken          string    `json:"verificationToken"`
	VerificationTokenExpiresAt time.Time `json:"verificationTokenExpiresAt"`
}

type UserVerificationEvent struct {
	Id     string `json:"id"`
	UserId string `json:"userId"`
	Token  string `json:"token"`
}

type PasswordResetEvent struct {
	Id     string `json:"id"`
	UserId string `json:"userId"`
	Token  string `json:"token"`
}
