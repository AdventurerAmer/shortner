package domain

import (
	"time"
)

type Topic string

const (
	ClicksTopic             Topic = "clicks"
	UserCreatedTopic        Topic = "users.created"
	UserVerificationTopic   Topic = "users.verification"
	UserPasswordResetTopcic Topic = "users.passwordReset"
	UserPendingDeleteTopcic Topic = "users.pendingDelete"
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
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type UserPasswordResetEvent struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type UserPendingDeleteEvent struct {
	Id string `json:"id"`
}
