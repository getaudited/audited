package main

import (
	"cmp"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/getaudited/audited/internal/domain"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/getaudited/audited/internal/adapters/psql"
	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/common/postgres"
	"github.com/getaudited/audited/internal/ports/http"
)

type Config struct {
	DatabaseURL       string   `envconfig:"DATABASE_URL"`
	HttpPort          int      `envconfig:"HTTP_PORT"`
	AllowedCorsOrigin []string `envconfig:"ALLOWED_CORS_ORIGIN"`
	DebugMode         bool     `envconfig:"DEBUG_MODE"`
	LogLevel          string   `envconfig:"LOG_LEVEL"`
	AmqpUrl           string   `envconfig:"AMQP_URL"`
	JWTPublicKey      string   `envconfig:"JWT_PUBLIC_KEY" required:"true"`
	JWTPrivateKey     string   `envconfig:"JWT_PRIVATE_KEY" required:"true"`
	AdminEmail        string   `envconfig:"ADMIN_EMAIL" required:"true"`
	AdminPassword     string   `envconfig:"ADMIN_PASSWORD" required:"true"`
}

type Service struct {
	config *Config
	logger *logs.Logger
}

func (s *Service) Run() error {
	logger := s.logger
	logger.Info("Kick starting service", "process_id", os.Getpid())

	config, err := s.loadEnvVariables()
	if err != nil {
		return err
	}
	s.config = config

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

	// Set up Admin user
	err = s.createAdminUserIfNotExists(ctx, db)
	if err != nil {
		return err
	}

	eventsRepository := psql.NewEventsPsqlRepository(db)
	sourcesRepository := psql.NewSourcesPsqlRepository(db)
	eventTypeRepository := psql.NewEventTypePsqlRepository(db)
	tokensRepository := psql.NewTokensPsqlRepository(db)
	usersRepository := psql.NewUsersPsqlRepository(db)
	jwtPrivateKey, err := s.parseJwtPrivateKey()
	if err != nil {
		return err
	}

	application := &app.App{
		Commands: app.Commands{
			CreateEventType: command.NewCreateEventTypeHandler(eventTypeRepository),
			DeleteEventType: command.NewDeleteEventTypeHandler(eventTypeRepository),
			CreateEvent:     command.NewCreateEventHandler(eventsRepository),
			CreateSource:    command.NewCreateSourceHandler(sourcesRepository),
			CreateToken:     command.NewCreateTokenHandler(tokensRepository),
			DeleteToken:     command.NewDeleteTokenHandler(tokensRepository),
			LogIn:           command.NewLogInHandler(usersRepository, jwtPrivateKey),
			CreateAdminUser: command.NewCreateAdminUserHandler(usersRepository),
		},
		Queries: app.Queries{
			EventTypes:        nil,
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypeRepository),
			Events:            query.NewAllEventsHandler(eventsRepository),
			EventByID:         nil,
			AllSources:        query.NewAllSourcesHandler(sourcesRepository),
			SourceByID:        query.NewSourceByIDHandler(sourcesRepository),
			AllTokens:         query.NewAllTokensHandler(tokensRepository),
			AllEventTypes:     query.NewAllEventTypesHandler(eventTypeRepository),
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

	g.Go(httpPort.Start)

	g.Go(func() error {
		<-ctx.Done()
		terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), time.Second*5)

		defer func() {
			terminationCtxCancel()
			cancel()
		}()

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
		return nil, fmt.Errorf("unable to load environment variables: %w", err)
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
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	parsedPublicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("the key provided is not an ECDS public key")
	}

	return parsedPublicKey, nil
}

func (s *Service) createAdminUserIfNotExists(ctx context.Context, db *sql.DB) error {
	usersRepository := psql.NewUsersPsqlRepository(db)

	email, err := domain.NewEmail(s.config.AdminEmail)
	if err != nil {
		return err
	}

	adminUser, err := usersRepository.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	if adminUser != nil {
		s.logger.Debug("Admin user exists")
		return nil
	}

	password, err := domain.NewPassword(s.config.AdminPassword)
	if err != nil {
		return err
	}

	user, err := domain.NewUser(email, password, domain.UserRoleAdmin)
	if err != nil {
		return err
	}

	handler := command.NewCreateAdminUserHandler(usersRepository)

	err = handler.Execute(ctx, command.CreateAdminUser{
		User: user,
	})
	if err != nil {
		return err
	}

	s.logger.Debug("Admin user set up successfully")

	return nil
}

func (s *Service) parseJwtPrivateKey() (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(strings.ReplaceAll(s.config.JWTPrivateKey, `\n`, "\n")))
	if block == nil {
		return nil, errors.New("failed to decode PEM block from 'JWT_PRIVATE_KEY'")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not ECDSA")
	}

	s.logger.Debug("JWT_PRIVATE_KEY loaded successfully")

	return ecKey, nil
}

func main() {
	logger := logs.New(cmp.Or(os.Getenv("LOG_LEVEL"), "INFO"))
	svc := &Service{
		logger: logger,
	}

	if err := svc.Run(); err != nil {
		logger.Error("service exited with an error", "error", err)
		os.Exit(1)
	}
}
