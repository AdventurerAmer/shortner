package config

import (
	"fmt"
	"net"
	"time"
)

type Infrastructure struct {
	Postgres   Postgres   `koanf:"postgres"`
	Cassandra  Cassandra  `koanf:"cassandra"`
	ClickHouse ClickHouse `koanf:"clickhouse"`
	Redis      Redis      `koanf:"redis"`
	Kafka      Kafka      `koanf:"kafka"`
}

type Postgres struct {
	Host               string        `koanf:"host" validate:"required,hostname"`
	Port               int           `koanf:"port" validate:"required,min=1024,max=65535"`
	Username           string        `koanf:"username" validate:"required,max=128"`
	Password           string        `koanf:"password" validate:"required,max=128"`
	Database           string        `koanf:"database" validate:"required"`
	SSLMode            string        `koanf:"sslMode" validate:"required,oneof=disable allow prefer require verify-ca verify-full"`
	MaxOpenConnections int           `koanf:"maxOpenConnections" validate:"required,min=1"`
	MaxIdelConnections int           `koanf:"maxIdelConnections" validate:"required,min=1"`
	MaxIdelTime        time.Duration `koanf:"maxIdelTime" validate:"required,min=1"`
}

func (p Postgres) Addr() string {
	return net.JoinHostPort(p.Host, fmt.Sprintf("%d", p.Port))
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
