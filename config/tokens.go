package config

import "time"

type Token struct {
	Name         string        `koanf:"name" validate:"required,max=128"`
	Secret       string        `koanf:"secret" validate:"required,max=128"`
	ExpiresAfter time.Duration `koanf:"expiresAfter" validate:"required,min=1m"`
}

type Tokens struct {
	UserVerification  Token `koanf:"userVerification"`
	UserPasswordReset Token `koanf:"userPasswordReset"`
}
