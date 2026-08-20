package config

import (
	"fmt"
	"time"
)

type ServicesConfig struct {
	Shortening  ServiceConfig `koanf:"shortening"`
	Redirecting ServiceConfig `koanf:"redirecting"`
	Analytics   ServiceConfig `koanf:"analytics"`
}

type ServiceConfig struct {
	Name                    string        `koanf:"name" validate:"required,min=1,max=128"`
	Port                    int           `koanf:"port" validate:"required,min=1,max=65535"`
	Version                 string        `koanf:"version" validate:"required,semver"`
	MaxHeaderBytes          int           `koanf:"maxHeaderBytes" validate:"required,min=1"`
	ReadHeaderTimeout       time.Duration `koanf:"readHeaderTimeout" validate:"required,min=1s"`
	ReadTimeout             time.Duration `koanf:"readTimeout" validate:"required,min=1s"`
	WriteTimeout            time.Duration `koanf:"writeTimeout" validate:"required,min=1s"`
	IdleTimeout             time.Duration `koanf:"idleTimeout" validate:"required,min=1s"`
	DefaultTimeout          time.Duration `koanf:"defaultTimeout" validate:"required,min=1s"`
	GracefulShutdownTimeout time.Duration `koanf:"gracefulShutdownTimeout" validate:"required,min=1s"`
	allowedOrigins          []string      `koanf:"allowedOrigins" validate:"required"`
}

func (srv *ServiceConfig) Address() string {
	// TODO: using http here
	return fmt.Sprintf("http://localhost:%d", srv.Port)
}

func setServiceDefaults(cfg *ServiceConfig) {
	if cfg.Name == "" {
		cfg.Name = "service"
	}

	if cfg.Port == 0 {
		cfg.Port = 3030
	}

	if cfg.Version == "" {
		cfg.Version = "0.0.1"
	}

	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = 1024 * 1024 // 1MB
	}

	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = time.Second
	}

	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = time.Second
	}

	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = time.Second
	}

	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = time.Minute
	}

	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = time.Second
	}

	if cfg.GracefulShutdownTimeout == 0 {
		cfg.GracefulShutdownTimeout = 10 * time.Second
	}

	if cfg.allowedOrigins == nil {
		cfg.allowedOrigins = []string{}
	}
}
