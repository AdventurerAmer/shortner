package config

import (
	"fmt"
	"net"
)

type Infrastructure struct {
	Cassandra      Cassandra  `koanf:"cassandra"`
	ClickHouse     ClickHouse `koanf:"clickhouse"`
	Redis          Redis      `koanf:"redis"`
	RedisAnalytics Redis      `koanf:"redisAnalytics"`
	Kafka          Kafka      `koanf:"kafka"`
}

type Cassandra struct {
	Host     string `koanf:"host" validate:"required,hostname"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Keyspace string `koanf:"keyspace" validate:"required,max=128"`
	Username string `koanf:"username" validate:"required,max=128"`
	Password string `koanf:"password" validate:"required,max=128"`
}

func (c Cassandra) Addr() string {
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))
}

type ClickHouse struct {
	Host     string `koanf:"host" validate:"required,hostname"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Database string `koanf:"database" validate:"required,max=128"`
	Username string `koanf:"username" validate:"required,max=128"`
	Password string `koanf:"password" validate:"required,max=128"`
}

func (ch ClickHouse) Addr() string {
	return net.JoinHostPort(ch.Host, fmt.Sprintf("%d", ch.Port))
}

type Redis struct {
	Host     string `koanf:"host" validate:"required,hostname"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Username string `koanf:"username" validate:"required,max=128"`
	Password string `koanf:"password" validate:"required,max=128"`
	Database *int   `koanf:"database" validate:"required,gte=0"`
}

func (r Redis) Addr() string {
	return net.JoinHostPort(r.Host, fmt.Sprintf("%d", r.Port))
}

type Kafka struct {
	Host     string `koanf:"host" validate:"required,hostname"`
	Port     int    `koanf:"port" validate:"required,min=1024,max=65535"`
	Username string `koanf:"username" validate:"required,max=128"`
	Password string `koanf:"password" validate:"required,max=128"`
}

func (k Kafka) Addr() string {
	return net.JoinHostPort(k.Host, fmt.Sprintf("%d", k.Port))
}
