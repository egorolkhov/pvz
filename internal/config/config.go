package config

import (
	"github.com/caarlos0/env/v8"
	"time"
)

type Config struct {
	ServerAddress     string        `env:"SERVER_ADDRESS" envDefault:":8080"`
	PrometheusAddress string        `env:"PROMETHEUS_ADDRESS" envDefault:":9000"`
	PostgresConn      string        `env:"POSTGRES_CONN" envDefault:"postgres://postgres:Qazxsw2200@localhost:5432/avito?sslmode=disable"`
	PostgresJDBCUrl   string        `env:"POSTGRES_JDBC_URL"`
	PostgresUser      string        `env:"POSTGRES_USERNAME"`
	PostgresPass      string        `env:"POSTGRES_PASSWORD"`
	PostgresHost      string        `env:"POSTGRES_HOST"`
	PostgresPort      string        `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresDB        string        `env:"POSTGRES_DATABASE"`
	JWTSecretKey      string        `env:"JWT_SECRET" envDefault:"qwerty"`
	TokenExp          time.Duration `env:"TOKEN_TTL"  envDefault:"15h"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
