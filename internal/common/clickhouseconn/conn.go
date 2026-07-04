package clickhouseconn

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func NewConnection(ctx context.Context, databaseURL string) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{strings.TrimPrefix(databaseURL, "clickhouse://")},
		Auth: clickhouse.Auth{
			Database: "default",  // TODO: Edit me
			Username: "default",  // TODO: Edit me
			Password: "password", // TODO: Edit me
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "an-example-go-client", Version: "0.1"},
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
