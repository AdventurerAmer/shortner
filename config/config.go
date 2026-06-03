package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/dotenv"
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

type AppConfig struct {
	Name        string `koanf:"name"`
	Environment Env    `koanf:"env"`
	Version     string `koanf:"version"`
}

type Config struct {
	App AppConfig `koanf:"app"`
}

func Load() (*Config, error) {
	delim := "."
	k := koanf.New(delim)

	envPrefix := "SHORTNER_"
	envDelim := "_"
	envParser := dotenv.ParserEnv(envPrefix, envDelim, func(s string) string {
		s = strings.TrimPrefix(s, envPrefix)
		s = strings.ReplaceAll(s, envDelim, delim)
		s = strings.ToLower(s)
		return s
	})
	if err := k.Load(file.Provider(".env.local"), envParser); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	var cfg Config
	unmarshalConf := koanf.UnmarshalConf{
		Tag: "koanf",
	}
	if err := k.UnmarshalWithConf("", &cfg, unmarshalConf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// setDefaults(&cfg)

	if err := validator.New().Struct(&cfg); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.App.Name == "" {
		cfg.App.Name = "Shortner"
	}
	if cfg.App.Environment == "" {
		cfg.App.Environment = EnvLocal
	}
	if cfg.App.Version == "" {
		cfg.App.Version = "0.1.0"
	}
}
