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

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/friendsofgo/errors"
	clickhouseadapter "github.com/getaudited/audited/internal/adapters/clickhouse"
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
	DatabaseURL           string   `envconfig:"ADT_DATABASE_URL"`
	ClickhouseDatabaseURL string   `envconfig:"ADT_CLICKHOUSE_DATABASE_URL"`
	HttpPort              int      `envconfig:"ADT_HTTP_PORT"`
	AllowedCorsOrigin     []string `envconfig:"ADT_ALLOWED_CORS_ORIGIN"`
	LogLevel              string   `envconfig:"ADT_LOG_LEVEL"`
	JWTPublicKey          string   `envconfig:"ADT_JWT_PUBLIC_KEY" required:"true"`
	JWTPrivateKey         string   `envconfig:"ADT_JWT_PRIVATE_KEY" required:"true"`
	AdminEmail            string   `envconfig:"ADT_ADMIN_EMAIL" required:"true"`
	AdminPassword         string   `envconfig:"ADT_ADMIN_PASSWORD" required:"true"`
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

	clickhouseConn, err := newClickhouseConnection(ctx, config.ClickhouseDatabaseURL)
	if err != nil {
		return err
	}

	// eventsRepo := psql.NewEventsPsqlRepository(db)
	eventsClickhouseRepo := clickhouseadapter.NewEventsClickhouseRepository(clickhouseConn)
	sourcesRepo := psql.NewSourcesPsqlRepository(db)
	eventTypeRepo := psql.NewEventTypePsqlRepository(db)
	tokensRepo := psql.NewTokensPsqlRepository(db)
	usersRepo := psql.NewUsersPsqlRepository(db)
	txProvider := psql.NewTxProvider(db, logger)
	jwtPrivateKey, err := s.parseJwtPrivateKey()
	if err != nil {
		return err
	}

	application := &app.App{
		Commands: app.Commands{
			CreateEventType:          command.NewCreateEventTypeHandler(eventTypeRepo),
			DeleteEventType:          command.NewDeleteEventTypeHandler(eventTypeRepo),
			CreateEventTypeVersion:   command.NewCreateEventTypeVersionHandler(txProvider),
			RollbackEventTypeVersion: command.NewRollbackEventTypeVersionHandler(eventTypeRepo),
			CreateEvent:              command.NewCreateEventHandler(eventsClickhouseRepo),
			CreateSource:             command.NewCreateSourceHandler(sourcesRepo),
			CreateToken:              command.NewCreateTokenHandler(tokensRepo),
			DeleteToken:              command.NewDeleteTokenHandler(tokensRepo),
			LogIn:                    command.NewLogInHandler(usersRepo, jwtPrivateKey),
			CreateAdminUser:          command.NewCreateAdminUserHandler(usersRepo),
		},
		Queries: app.Queries{
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypeRepo),
			Events:            query.NewAllEventsHandler(eventsClickhouseRepo),
			EventByID:         query.NewEventByIDHandler(eventsClickhouseRepo),
			AllSources:        query.NewAllSourcesHandler(sourcesRepo),
			SourceByID:        query.NewSourceByIDHandler(sourcesRepo),
			AllTokens:         query.NewAllTokensHandler(tokensRepo),
			AllEventTypes:     query.NewAllEventTypesHandler(eventTypeRepo),
			UserProfile:       query.NewUserProfileHandler(usersRepo),
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
		return nil, errors.New("failed to decode PEM block from 'ADT_JWT_PRIVATE_KEY'")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not ECDSA")
	}

	s.logger.Debug("ADT_JWT_PRIVATE_KEY loaded successfully")

	return ecKey, nil
}

func main() {
	logger := logs.New(cmp.Or(os.Getenv("ADT_LOG_LEVEL"), "INFO"))
	svc := &Service{
		logger: logger,
	}

	if err := svc.Run(); err != nil {
		logger.Error("service exited with an error", "error", err)
		os.Exit(1)
	}
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
