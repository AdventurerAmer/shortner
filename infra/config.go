package infra

import (
	"fmt"
	"time"
)

type Config struct {
	startupTimeout  time.Duration
	shutdownTimeout time.Duration
}

type Option = func(cfg *Config) error

func WithStartupTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("startupTimeout must not be zero")
		}
		cfg.startupTimeout = timeout
		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("shutdownTimeout must not be zero")
		}
		cfg.shutdownTimeout = timeout
		return nil
	}
}
