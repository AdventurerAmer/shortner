package config

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type InfrastructureConfig struct {
	Database       CassandraDatabaseConfig `koanf:"database"`
	Redis          RedisConfig             `koanf:"redis"`
	RedisAnalytics RedisConfig             `koanf:"redisAnalytics"`
	Kafka          KafkaConfig             `koanf:"kafka"`
}

type CassandraDatabaseConfig struct {
	Host        string        `koanf:"host" validate:"required,hostname"`
	Port        int           `koanf:"port" validate:"required,min=1,max=65535"`
	Keyspace    string        `koanf:"keyspace" validate:"required,min=1"`
	ConnTimeout time.Duration `koanf:"connTimeout" validate:"required,min=1s"`
}

type RedisConfig struct {
	Address  string `koanf:"address" validate:"required,url"`
	Username string `koanf:"username" validate:"omitempty,min=1"`
	Password string `koanf:"password" validate:"required,min=1"`
	Database int    `koanf:"database" validate:"required,min=0"`
}

type Redis struct {
	Client *redis.Client
}

type KafkaConfig struct {
	Host string `koanf:"host" validate:"required,hostname"`
	Port int    `koanf:"port" validate:"required,min=1,max=65535"`
}
