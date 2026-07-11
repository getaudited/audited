package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ClickhouseHosts                 []string `envconfig:"ADT_CLICKHOUSE_HOSTS"`
	ClickhouseDbName                string   `envconfig:"ADT_CLICKHOUSE_DBNAME"`
	ClickhouseUsername              string   `envconfig:"ADT_CLICKHOUSE_USERNAME"`
	ClickhousePassword              string   `envconfig:"ADT_CLICKHOUSE_PASSWORD"`
	ClickhouseTlsEnabled            bool     `envconfig:"ADT_CLICKHOUSE_TLS_ENABLED"`
	ClickhouseTlsInsecureSkipVerify bool     `envconfig:"ADT_CLICKHOUSE_TLS_INSECURE_SKIP_VERIFY"`

	DatabaseURL string `envconfig:"ADT_DATABASE_URL"`

	HttpPort          int      `envconfig:"ADT_HTTP_PORT"`
	AllowedCorsOrigin []string `envconfig:"ADT_ALLOWED_CORS_ORIGIN"`

	LogLevel string `envconfig:"ADT_LOG_LEVEL"`

	JWTPublicKey  string `envconfig:"ADT_JWT_PUBLIC_KEY"`
	JWTPrivateKey string `envconfig:"ADT_JWT_PRIVATE_KEY"`
	JWTSecret     string `envconfig:"ADT_JWT_SECRET"`

	AdminEmail    string `envconfig:"ADT_ADMIN_EMAIL" required:"true"`
	AdminPassword string `envconfig:"ADT_ADMIN_PASSWORD" required:"true"`
}

func (c *Config) JwtKeysSet() bool {
	return strings.TrimSpace(c.JWTPublicKey) != "" && strings.TrimSpace(c.JWTPrivateKey) != ""
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.JWTSecret) == "" && !c.JwtKeysSet() {
		return fmt.Errorf("set ADT_JWT_SECRET or both ADT_JWT_PUBLIC_KEY and ADT_JWT_PRIVATE_KEY")
	}

	return nil
}

func New() (*Config, error) {
	envs := &Config{}

	err := envconfig.Process("", envs)
	if err != nil {
		return nil, fmt.Errorf("unable to load environment variables: %w", err)
	}

	err = envs.validate()
	if err != nil {
		return nil, err
	}

	return envs, nil
}
