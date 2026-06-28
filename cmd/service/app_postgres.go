package main

import (
	"context"
	"crypto/ecdsa"

	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/common/postgres"
)

func NewPostgresApp(
	ctx context.Context,
	logger *logs.Logger,
	jwtPrivateKey *ecdsa.PrivateKey,
	config *Config,
) (*app.App, Closer, error) {
	db, err := postgres.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	err = postgres.ApplyMigrations(db, "misc/sql/migrations")
	if err != nil {
		return nil, nil, err
	}

	eventsRepo := psql.NewEventsPsqlRepository(db)
	sourcesRepo := psql.NewSourcesPsqlRepository(db)
	eventTypeRepo := psql.NewEventTypePsqlRepository(db)
	tokensRepo := psql.NewTokensPsqlRepository(db)
	usersRepo := psql.NewUsersPsqlRepository(db)
	txProvider := psql.NewTxProvider(db, logger)

	return &app.App{
		Commands: app.Commands{
			CreateEventType:          command.NewCreateEventTypeHandler(eventTypeRepo),
			DeleteEventType:          command.NewDeleteEventTypeHandler(eventTypeRepo),
			CreateEventTypeVersion:   command.NewCreateEventTypeVersionHandler(txProvider),
			RollbackEventTypeVersion: command.NewRollbackEventTypeVersionHandler(eventTypeRepo),
			CreateEvent:              command.NewCreateEventHandler(eventsRepo),
			CreateSource:             command.NewCreateSourceHandler(sourcesRepo),
			CreateToken:              command.NewCreateTokenHandler(tokensRepo),
			DeleteToken:              command.NewDeleteTokenHandler(tokensRepo),
			LogIn:                    command.NewLogInHandler(usersRepo, jwtPrivateKey),
			CreateAdminUser:          command.NewCreateAdminUserHandler(usersRepo),
		},
		Queries: app.Queries{
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypeRepo),
			Events:            query.NewAllEventsHandler(eventsRepo),
			EventByID:         query.NewEventByIDHandler(eventsRepo),
			AllSources:        query.NewAllSourcesHandler(sourcesRepo),
			SourceByID:        query.NewSourceByIDHandler(sourcesRepo),
			AllTokens:         query.NewAllTokensHandler(tokensRepo),
			AllEventTypes:     query.NewAllEventTypesHandler(eventTypeRepo),
			UserProfile:       query.NewUserProfileHandler(usersRepo),
		},
	}, db, nil
}
