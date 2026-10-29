package domain

import "time"

type TokenType string

const (
	TokenTypeVerification  TokenType = "verification"
	TokenTypePasswordReset TokenType = "passwordReset"
)

type Token struct {
	Id        string     `json:"id"`
	Type      TokenType  `json:"type"`
	UserId    string     `json:"userId"`
	Hash      string     `json:"hash"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt"`
	IsRevoked bool       `json:"isRevoked"`
}
