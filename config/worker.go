package config

import (
	"time"
)

type WorkersConfig struct {
	Clicks        WorkerConfig `koanf:"clicks"`
	ClicksBatcher WorkerConfig `koanf:"clicksBatcher"`
}

type WorkerConfig struct {
	Name                    string        `koanf:"name" validate:"required,min=1,max=128"`
	Version                 string        `koanf:"version" validate:"required,semver"`
	Group                   string        `koanf:"group" validate:"required,min=1"`
	Topic                   string        `koanf:"topic" validate:"required,oneof=clicks clicksBatch"`
	DefaultTimeout          time.Duration `koanf:"defaultTimeout" validate:"required,min=1s"`
	GracefulShutdownTimeout time.Duration `koanf:"gracefulShutdownTimeout" validate:"required,min=1s"`
	AckTimeout              time.Duration `koanf:"AckTimeout" validate:"required,min=1s"`
	AckRetries              int           `koanf:"AckRetries" validate:"required,min=1,max=16"`
}

func setWorkerDefaults(cfg *WorkerConfig) {
	if cfg.Name == "" {
		cfg.Name = "worker"
	}
	if cfg.Version == "" {
		cfg.Version = "0.0.1"
	}

	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = time.Second
	}

	if cfg.AckTimeout == 0 {
		cfg.AckTimeout = time.Second
	}

	if cfg.AckRetries == 0 {
		cfg.AckRetries = 10
	}

	if cfg.GracefulShutdownTimeout == 0 {
		cfg.GracefulShutdownTimeout = 10 * time.Second
	}
}
