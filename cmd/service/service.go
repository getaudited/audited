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

	"github.com/tachyonhqdev/webhooks/internal/adapters/psql"
	"github.com/tachyonhqdev/webhooks/internal/app"
	"github.com/tachyonhqdev/webhooks/internal/app/command"
	"github.com/tachyonhqdev/webhooks/internal/common/logs"
	messaginglib "github.com/tachyonhqdev/webhooks/internal/common/messaging"
	"github.com/tachyonhqdev/webhooks/internal/common/postgres"
	"github.com/tachyonhqdev/webhooks/internal/ports/amqp"
	"github.com/tachyonhqdev/webhooks/internal/ports/http"
)

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

	db, err := postgres.Connect(config.DatabaseURL)
	if err != nil {
		return err
	}

	err = postgres.ApplyMigrations(db, "misc/sql/migrations")
	if err != nil {
		return err
	}

	messaging, err := messaginglib.NewMessaging(db, config.AmqpUrl, logger)
	if err != nil {
		return err
	}

	eventTypeRepository := psql.NewEventTypePsqlRepository(db)

	application := &app.App{
		Commands: app.Commands{
			CreateEventType: command.NewCreateEventTypeHandler(eventTypeRepository),
		},
		Queries: app.Queries{},
	}

	httpPort, err := http.NewServer(http.Config{
		Application:        application,
		Port:               config.HttpPort,
		AllowedCorsOrigin:  config.AllowedCorsOrigin,
		Logger:             logger,
		IsDebug:            config.DebugMode,
		Ctx:                ctx,
		JwtSecret:          "",
		WebFrontendEnabled: false,
	})
	if err != nil {
		return err
	}

	err = messaging.CommandProcessor().AddHandlers(amqp.NewCommandHandlers(application)...)
	if err != nil {
		return fmt.Errorf("error registering command handlers")
	}

	g.Go(func() error {
		return messaging.CommandForwarder().Run(ctx)
	})

	g.Go(func() error {
		return messaging.Router().Run(ctx)
	})

	g.Go(func() error {
		<-messaging.Router().Running()
		return httpPort.Start()
	})

	g.Go(func() error {
		<-ctx.Done()
		terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), time.Second*5)

		defer func() {
			terminationCtxCancel()
			cancel()
		}()

		err = messaging.Close()
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
