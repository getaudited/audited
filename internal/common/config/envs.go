package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ClickhouseHosts    []string `envconfig:"ADT_CLICKHOUSE_HOSTS"`
	ClickhouseDbName   string   `envconfig:"ADT_CLICKHOUSE_DBNAME"`
	ClickhouseUsername string   `envconfig:"ADT_CLICKHOUSE_USERNAME"`
	ClickhousePassword string   `envconfig:"ADT_CLICKHOUSE_PASSWORD"`
	DatabaseURL        string   `envconfig:"ADT_DATABASE_URL"`
	HttpPort           int      `envconfig:"ADT_HTTP_PORT"`
	AllowedCorsOrigin  []string `envconfig:"ADT_ALLOWED_CORS_ORIGIN"`
	LogLevel           string   `envconfig:"ADT_LOG_LEVEL"`
	JWTPublicKey       string   `envconfig:"ADT_JWT_PUBLIC_KEY" required:"true"`
	JWTPrivateKey      string   `envconfig:"ADT_JWT_PRIVATE_KEY" required:"true"`
	AdminEmail         string   `envconfig:"ADT_ADMIN_EMAIL" required:"true"`
	AdminPassword      string   `envconfig:"ADT_ADMIN_PASSWORD" required:"true"`
}

func New() (*Config, error) {
	envs := &Config{}

	err := envconfig.Process("", envs)
	if err != nil {
		return nil, fmt.Errorf("unable to load environment variables: %w", err)
	}

	return envs, nil
}
