package config

import "time"

type InfrastructureConfig struct {
	Database CassandraDatabaseConfig `koanf:"database"`
}

type CassandraDatabaseConfig struct {
	Host        string        `koanf:"host" validate:"required,hostname"`
	Port        int           `koanf:"port" validate:"required,min=1,max=65535"`
	Keyspace    string        `koanf:"keyspace" validate:"required,min=1"`
	ConnTimeout time.Duration `koanf:"connTimeout" validate:"required,min=1s"`
}
