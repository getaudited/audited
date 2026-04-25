package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/firminochangani/audited/internal/adapters/psql"
	"github.com/firminochangani/audited/internal/app"
	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/common/logs"
	"github.com/firminochangani/audited/internal/common/postgres"
	"github.com/firminochangani/audited/internal/ports/http"
)

func main() {
	logger := logs.New()
	svc := &Service{
		logger: logger,
	}

	if err := svc.Run(); err != nil {
		logger.Error("service exited with an error", "error", err)
		os.Exit(1)
	}
}

type Config struct {
	DatabaseURL       string   `envconfig:"DATABASE_URL"`
	HttpPort          int      `envconfig:"HTTP_PORT"`
	AllowedCorsOrigin []string `envconfig:"ALLOWED_CORS_ORIGIN"`
	DebugMode         bool     `envconfig:"DEBUG_MODE"`
	AmqpUrl           string   `envconfig:"AMQP_URL"`
}

type Service struct {
	logger *logs.Logger
}

func (s *Service) Run() error {
	logger := s.logger
	logger.Info("Kick starting service", "process_id", os.Getpid())

	config, err := s.loadEnvVariables()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	g, ctx := errgroup.WithContext(ctx)

	db, err := postgres.Connect(ctx, config.DatabaseURL)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	err = postgres.ApplyMigrations(db, "misc/sql/migrations")
	if err != nil {
		return err
	}

	eventsRepository := psql.NewEventsPsqlRepository(db)
	sourcesRepository := psql.NewSourcesPsqlRepository(db)
	eventTypeRepository := psql.NewEventTypePsqlRepository(db)
	tokensRepository := psql.NewTokensPsqlRepository(db)

	application := &app.App{
		Commands: app.Commands{
			CreateEventType: command.NewCreateEventTypeHandler(eventTypeRepository),
			CreateEvent:     command.NewCreateEventHandler(eventsRepository),
			CreateSource:    command.NewCreateSourceHandler(sourcesRepository),
			CreateToken:     command.NewCreateTokenHandler(tokensRepository),
		},
		Queries: app.Queries{
			EventTypes:        nil,
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypeRepository),
			Events:            nil,
			EventByID:         nil,
			AllSources:        query.NewAllSourcesHandler(sourcesRepository),
		},
	}

	httpPort, err := http.NewServer(http.Config{
		Application:       application,
		Port:              config.HttpPort,
		AllowedCorsOrigin: config.AllowedCorsOrigin,
		Logger:            logger,
		IsDebug:           config.DebugMode,
		Context:           ctx,
	})
	if err != nil {
		return err
	}

	g.Go(func() error {
		return httpPort.Start()
	})

	g.Go(func() error {
		<-ctx.Done()
		terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), time.Second*5)

		defer func() {
			terminationCtxCancel()
			cancel()
		}()

		if err != nil {
			return fmt.Errorf("error closing messaging: %v", err)
		}

		err = httpPort.Stop(terminationCtx)
		if err != nil {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (s *Service) loadEnvVariables() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("unable to load environment variables: %v", err)
	}

	return config, nil
}
