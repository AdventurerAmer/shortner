package worker

import (
	"fmt"
	"time"
)

type Config struct {
	gracefulShutdownTimeout time.Duration
	ackTimeout              time.Duration
	ackRetries              int
}

type Option func(cfg *Config) error

func WithGracefulShutdownTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("gracefulShutdownTimeout must not be zero")
		}
		cfg.gracefulShutdownTimeout = timeout
		return nil
	}
}

func WithAckTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("ackTimeout must not be zero")
		}
		cfg.ackTimeout = timeout
		return nil
	}
}

func WithAckRetries(retries int) Option {
	return func(cfg *Config) error {
		if retries <= 0 {
			return fmt.Errorf("ackRetries must be positive")
		}
		cfg.ackRetries = retries
		return nil
	}
}
