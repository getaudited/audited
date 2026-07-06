package clickhouseconn

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Config struct {
	Version  string
	Hosts    []string
	Database string
	Username string
	Password string
}

func NewConnection(ctx context.Context, config Config) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
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
		/*TLS: &tls.Config{
			InsecureSkipVerify: true,
		},*/
	})

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
