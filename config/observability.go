package config

import "time"

type ObservabilityConfig struct {
	Logging LoggingConfig `koanf:"logging"`
	// Tracing      TracingConfig      `koanf:"tracing"`
	HealthChecks HealthChecksConfig `koanf:"healthChecks"`
}

type LoggingConfig struct {
	Level     string `koanf:"level" validate:"oneof=debug info warn error"`
	Format    string `koanf:"format" validate:"oneof=json text"`
	AddSource *bool  `koanf:"addSource"`
}

type TracingConfig struct {
	Enabled  bool   `koanf:"enabled"`
	Endpoint string `koanf:"endpoint" validate:"required,url"`
}

type HealthChecksConfig struct {
	Enabled  bool          `koanf:"enabled"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Checks   []string      `koanf:"checks"`
}
