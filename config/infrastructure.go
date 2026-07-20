package config

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type InfrastructureConfig struct {
	Cassandra      CassandraConfig  `koanf:"cassandra"`
	ClickHouse     ClickHouseConfig `koanf:"clickhouse"`
	Redis          RedisConfig      `koanf:"redis"`
	RedisAnalytics RedisConfig      `koanf:"redisAnalytics"`
	Kafka          KafkaConfig      `koanf:"kafka"`
}

type CassandraConfig struct {
	Host        string        `koanf:"host" validate:"required,hostname"`
	Port        int           `koanf:"port" validate:"required,min=1,max=65535"`
	Keyspace    string        `koanf:"keyspace" validate:"required,min=1"`
	ConnTimeout time.Duration `koanf:"connTimeout" validate:"required,min=1s"`
}

type ClickHouseConfig struct {
	Host     string `koanf:"host" validate:"required,hostname"`
	Port     int    `koanf:"port" validate:"required,min=1,max=65535"`
	Database string `koanf:"database" validate:"required,min=1"`
	Username string `koanf:"username" validate:"required,min=1"`
	Password string `koanf:"password" validate:"required,min=1"`
}

type RedisConfig struct {
	Address  string `koanf:"address" validate:"required,url"`
	Username string `koanf:"username" validate:"required,min=1"`
	Password string `koanf:"password" validate:"required,min=1"`
	Database *int   `koanf:"database" validate:"required,gte=0"`
}

type Redis struct {
	Client *redis.Client
}

type KafkaConfig struct {
	Host string `koanf:"host" validate:"required,hostname"`
	Port int    `koanf:"port" validate:"required,min=1,max=65535"`
}
