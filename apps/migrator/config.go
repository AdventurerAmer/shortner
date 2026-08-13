package migrator

import (
	"fmt"
	"time"
)

type Config struct {
	queryTimeout time.Duration
}

type Option = func(cfg *Config) error

func WithQueryTimeout(timeout time.Duration) Option {
	return func(cfg *Config) error {
		if timeout == 0 {
			return fmt.Errorf("queryTimeout must not be zero")
		}
		cfg.queryTimeout = timeout
		return nil
	}
}
