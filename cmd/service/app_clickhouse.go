package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	clickhouseadapter "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
)

type Closer interface {
	Close() error
}

func NewClickhouseApp(
	ctx context.Context,
	config *Config,
	jwtPrivateKey *ecdsa.PrivateKey,
) (*app.App, Closer, error) {
	conn, err := newClickhouseConnection(ctx, config.ClickhouseDatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	tokensRepo := clickhouseadapter.NewTokenChRepository(conn)
	eventsRepo := clickhouseadapter.NewEventsClickhouseRepository(conn)
	sourcesRepo := clickhouseadapter.NewSourcesClickhouseRepository(conn)
	usersRepo := clickhouseadapter.NewUsersClickhouseRepository(conn)
	eventTypesRepo := clickhouseadapter.NewEventTypesClickhouseRepository(conn)
	shallowTxProvider := clickhouseadapter.NewShallowTxProvider(conn)

	return &app.App{
		Commands: app.Commands{
			CreateEventType:          command.NewCreateEventTypeHandler(eventTypesRepo),
			DeleteEventType:          command.NewDeleteEventTypeHandler(eventTypesRepo),
			CreateEventTypeVersion:   command.NewCreateEventTypeVersionHandler(shallowTxProvider),
			RollbackEventTypeVersion: command.NewRollbackEventTypeVersionHandler(eventTypesRepo),

			CreateEvent: command.NewCreateEventHandler(eventsRepo),

			CreateSource: command.NewCreateSourceHandler(sourcesRepo),

			CreateToken: command.NewCreateTokenHandler(tokensRepo),
			DeleteToken: command.NewDeleteTokenHandler(tokensRepo),

			LogIn:           command.NewLogInHandler(usersRepo, jwtPrivateKey),
			CreateAdminUser: command.NewCreateAdminUserHandler(usersRepo),
		},
		Queries: app.Queries{
			Events:    query.NewAllEventsHandler(eventsRepo),
			EventByID: query.NewEventByIDHandler(eventsRepo),

			AllEventTypes:     query.NewAllEventTypesHandler(eventTypesRepo),
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypesRepo),

			AllSources: query.NewAllSourcesHandler(sourcesRepo),
			SourceByID: query.NewSourceByIDHandler(sourcesRepo),

			AllTokens: query.NewAllTokensHandler(tokensRepo),

			UserProfile: query.NewUserProfileHandler(usersRepo),
		},
	}, conn, nil
}

func newClickhouseConnection(ctx context.Context, databaseURL string) (clickhouse.Conn, error) {
	var conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{strings.TrimPrefix(databaseURL, "clickhouse://")},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "password",
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "an-example-go-client", Version: "0.1"},
			},
		},
		Debugf: func(format string, v ...interface{}) {
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
