package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/friendsofgo/errors"
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
	JWTPublicKey      string   `envconfig:"JWT_PUBLIC_KEY" required:"true"`
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
			DeleteEventType: command.NewDeleteEventTypeHandler(eventTypeRepository),
			CreateEvent:     command.NewCreateEventHandler(eventsRepository),
			CreateSource:    command.NewCreateSourceHandler(sourcesRepository),
			CreateToken:     command.NewCreateTokenHandler(tokensRepository),
			DeleteToken:     command.NewDeleteTokenHandler(tokensRepository),
		},
		Queries: app.Queries{
			EventTypes:        nil,
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypeRepository),
			Events:            query.NewAllEventsHandler(eventsRepository),
			EventByID:         nil,
			AllSources:        query.NewAllSourcesHandler(sourcesRepository),
		},
	}

	jwtPublicKey, err := s.parsePublicKey(config.JWTPublicKey)
	if err != nil {
		return err
	}

	httpPort, err := http.NewServer(http.Config{
		Application:       application,
		Port:              config.HttpPort,
		AllowedCorsOrigin: config.AllowedCorsOrigin,
		Logger:            logger,
		IsDebug:           config.DebugMode,
		Context:           ctx,
		JwtPublicKey:      jwtPublicKey,
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

func (s *Service) parsePublicKey(content string) (*ecdsa.PublicKey, error) {
	content = strings.ReplaceAll(content, `\n`, "\n")

	block, _ := pem.Decode([]byte(content))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("error decoding public key's PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	parsedPublicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("the key provided is not an ECDS public key")
	}

	return parsedPublicKey, nil
}
