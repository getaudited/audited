package clickhouseconn

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/pressly/goose/v3"
)

type Config struct {
	Version               string
	Hosts                 []string
	Database              string
	Username              string
	Password              string
	TlsEnabled            bool
	TlsInsecureSkipVerify bool
}

func NewConnection(ctx context.Context, config Config) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(mapConfigToOptions(config))
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}

		return nil, err
	}

	return conn, nil
}

func ApplyMigrations(ctx context.Context, cfg Config, dir string, logger *logs.Logger) error {
	db := clickhouse.OpenDB(mapConfigToOptions(cfg))
	defer func() { _ = db.Close() }()

	err := goose.SetDialect("clickhouse")
	if err != nil {
		return fmt.Errorf("migrations: error setting the dialect : %w", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("migrations: error pinging db: %w", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("migrations: error applying migration: %w", err)
	}

	logger.Info("Clickhouse migration applied successfully")

	return nil
}

func mapConfigToOptions(config Config) *clickhouse.Options {
	opts := &clickhouse.Options{
		Addr: config.Hosts,
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "audited", Version: config.Version},
			},
		},
		Debugf: func(format string, v ...any) {
			fmt.Printf(format, v)
		},
	}

	if config.TlsEnabled {
		opts.TLS = &tls.Config{
			InsecureSkipVerify: config.TlsInsecureSkipVerify,
		}
	}

	return opts
}
