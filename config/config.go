package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/env/v2"
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
	if err := godotenv.Load(".env.local"); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	delim := "."
	k := koanf.New(delim)

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
