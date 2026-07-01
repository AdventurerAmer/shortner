package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Env string

func (env Env) String() string {
	return string(env)
}

const (
	EnvLocal   Env = "local"
	EnvStaging Env = "staging"
	EnvProd    Env = "production"
)

var Envs = []Env{EnvLocal, EnvStaging, EnvProd}

type Config struct {
	Env            Env                  `koanf:"env" validate:"required,oneof=local staging production"`
	App            AppConfig            `koanf:"app"`
	Observability  ObservabilityConfig  `koanf:"observability"`
	Services       ServicesConfig       `koanf:"services"`
	Infrastructure InfrastructureConfig `koanf:"infrastructure"`
}

type AppConfig struct {
	Name    string `koanf:"name" validate:"required,min=1,max=128"`
	Version string `koanf:"version" validate:"required,semver"`
}

type ServiceConfig struct {
	Name                    string        `koanf:"name" validate:"required,min=1,max=128"`
	Port                    int           `koanf:"port" validate:"required,min=1,max=65535"`
	MaxHeaderBytes          int           `koanf:"maxHeaderBytes" validate:"required,min=1"`
	ReadHeaderTimeout       time.Duration `koanf:"ReadHeaderTimeout" validate:"required,min=1s"`
	ReadTimeout             time.Duration `koanf:"readTimeout" validate:"required,min=1s"`
	WriteTimeout            time.Duration `koanf:"writeTimeout" validate:"required,min=1s"`
	IdleTimeout             time.Duration `koanf:"idleTimeout" validate:"required,min=1s"`
	DefaultTimeout          time.Duration `koanf:"defaultTimeout" validate:"required,min=1s"`
	GracefulShutdownTimeout time.Duration `koanf:"gracefulShutdownTimeout" validate:"required,min=1s"`
	allowedOrigins          []string      `koanf:"allowedOrigins" validate:"required"`
}

type ServicesConfig struct {
	Shortening  ServiceConfig `koanf:"shortening"`
	Redirecting ServiceConfig `koanf:"redirecting"`
	Analytics   ServiceConfig `koanf:"analytics"`
}

func Load() (*Config, error) {
	delim := "."
	k := koanf.New(delim)

	// Load config.yaml first
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to load config.yaml: %w", err)
	}

	// Override config from .env
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	envPrefix := "SHORTNER_"
	envDelim := "_"

	envOpt := env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.TrimPrefix(k, envPrefix)
			k = strings.ReplaceAll(k, envDelim, delim)
			k = strings.ToLower(k)
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}
			return k, v
		},
	}
	err := k.Load(env.Provider(".", envOpt), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %w", err)
	}

	unmarshalConf := koanf.UnmarshalConf{
		Tag: "koanf",
	}
	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, unmarshalConf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	setDefaults(&cfg)

	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Env == "" {
		cfg.Env = EnvLocal
	}

	if cfg.App.Name == "" {
		cfg.App.Name = "Shortner"
	}

	if cfg.App.Version == "" {
		cfg.App.Version = "0.1.0"
	}

	if cfg.Observability.Logging.Level == "" {
		if cfg.Env == EnvLocal {
			cfg.Observability.Logging.Level = "debug"
		} else {
			cfg.Observability.Logging.Level = "info"
		}
	}

	if cfg.Observability.Logging.Format == "" {
		if cfg.Env == EnvLocal {
			cfg.Observability.Logging.Format = "text"
		} else {
			cfg.Observability.Logging.Format = "json"
		}
	}

	if cfg.Observability.Logging.AddSource == nil {
		addSource := (cfg.Env == EnvLocal || cfg.Env == EnvStaging)
		cfg.Observability.Logging.AddSource = &addSource
	}

	if cfg.Observability.HealthChecks.Interval == 0 {
		cfg.Observability.HealthChecks.Interval = 30 * time.Second
	}

	if cfg.Observability.HealthChecks.Timeout == 0 {
		cfg.Observability.HealthChecks.Timeout = 5 * time.Second
	}

	setServiceDefaults(&cfg.Services.Shortening)
	setServiceDefaults(&cfg.Services.Redirecting)
	setServiceDefaults(&cfg.Services.Analytics)
}

func setServiceDefaults(cfg *ServiceConfig) {
	if cfg.Name == "" {
		cfg.Name = "service"
	}

	if cfg.Port == 0 {
		cfg.Port = 3030
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
