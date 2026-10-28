package config

import "time"

type Auth struct {
	Secret            string        `koanf:"secret" validate:"required"`
	TokenExpiresAfter time.Duration `koanf:"tokenExpiresAfter" validate:"required,min=15m"`
}
