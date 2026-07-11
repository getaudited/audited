package main

import (
	"cmp"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	clickhouseadapter "github.com/getaudited/audited/internal/adapters/clickhouse"
	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/common/clickhouseconn"
	"github.com/getaudited/audited/internal/common/config"
	"github.com/getaudited/audited/internal/common/logs"
	"github.com/getaudited/audited/internal/domain"
	"github.com/getaudited/audited/internal/ports/http"
)

var Version = "development"

type Service struct {
	config *config.Config
	logger *logs.Logger
}

func (s *Service) Run() error {
	logger := s.logger
	logger.Info("Kick starting service", "process_id", os.Getpid())

	cfg, err := config.New()
	if err != nil {
		return err
	}
	s.config = cfg

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	g, ctx := errgroup.WithContext(ctx)

	var jwtPublicKey *ecdsa.PublicKey
	var jwtPrivateKey *ecdsa.PrivateKey

	if cfg.JwtKeysSet() {
		jwtPrivateKey, err = s.parseJwtPrivateKey()
		if err != nil {
			return err
		}

		jwtPublicKey, err = s.parsePublicKey(cfg.JWTPublicKey)
		if err != nil {
			return err
		}
	}

	clickhousecfg := clickhouseconn.Config{
		Version:               Version,
		Hosts:                 cfg.ClickhouseHosts,
		Database:              cfg.ClickhouseDbName,
		Username:              cfg.ClickhouseUsername,
		Password:              cfg.ClickhousePassword,
		TlsEnabled:            cfg.ClickhouseTlsEnabled,
		TlsInsecureSkipVerify: cfg.ClickhouseTlsInsecureSkipVerify,
	}

	conn, err := clickhouseconn.NewConnection(ctx, clickhousecfg)
	if err != nil {
		return err
	}

	err = clickhouseconn.ApplyMigrations(ctx, clickhousecfg, "misc/clickhouse", logger)
	if err != nil {
		return err
	}

	tokensRepo := clickhouseadapter.NewTokenChRepository(conn)
	eventsRepo := clickhouseadapter.NewEventsClickhouseRepository(conn)
	sourcesRepo := clickhouseadapter.NewSourcesClickhouseRepository(conn)
	usersRepo := clickhouseadapter.NewUsersClickhouseRepository(conn)
	eventTypesRepo := clickhouseadapter.NewEventTypesClickhouseRepository(conn)
	shallowTxProvider := clickhouseadapter.NewShallowTxProvider(conn)

	application := &app.App{
		Commands: app.Commands{
			CreateEventType:          command.NewCreateEventTypeHandler(eventTypesRepo),
			DeleteEventType:          command.NewDeleteEventTypeHandler(eventTypesRepo),
			CreateEventTypeVersion:   command.NewCreateEventTypeVersionHandler(shallowTxProvider),
			RollbackEventTypeVersion: command.NewRollbackEventTypeVersionHandler(eventTypesRepo),

			CreateEvent: command.NewCreateEventHandler(eventsRepo),

			CreateSource: command.NewCreateSourceHandler(sourcesRepo),

			CreateToken: command.NewCreateTokenHandler(tokensRepo),
			DeleteToken: command.NewDeleteTokenHandler(tokensRepo),

			LogIn:           command.NewLogInHandler(usersRepo, jwtPrivateKey, cfg.JWTSecret),
			CreateAdminUser: command.NewCreateAdminUserHandler(usersRepo),
		},
		Queries: app.Queries{
			Events:    query.NewAllEventsHandler(eventsRepo),
			EventByID: query.NewEventByIDHandler(eventsRepo),

			AllEventTypes:     query.NewAllEventTypesHandler(eventTypesRepo),
			EventTypeVersions: query.NewEventTypeVersionsHandler(eventTypesRepo),
			EventTypeByAction: query.NewEventTypeByActionHandler(eventTypesRepo),

			AllSources: query.NewAllSourcesHandler(sourcesRepo),
			SourceByID: query.NewSourceByIDHandler(sourcesRepo),

			AllTokens: query.NewAllTokensHandler(tokensRepo),

			UserProfile: query.NewUserProfileHandler(usersRepo),
		},
	}

	// Set up Admin user
	err = s.createAdminUserIfNotExists(ctx, application)
	if err != nil {
		return err
	}

	httpPort, err := http.NewServer(http.Config{
		Context:           ctx,
		Logger:            logger,
		Application:       application,
		JwtPublicKey:      jwtPublicKey,
		JwtSecret:         cfg.JWTSecret,
		Port:              cfg.HttpPort,
		AllowedCorsOrigin: cfg.AllowedCorsOrigin,
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

		return conn.Close()
	})

	return g.Wait()
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

func (s *Service) createAdminUserIfNotExists(ctx context.Context, app *app.App) error {
	err := app.Commands.CreateAdminUser.Execute(ctx, command.CreateAdminUser{
		Email:    s.config.AdminEmail,
		Password: s.config.AdminPassword,
	})
	if errors.Is(err, domain.ErrUserExists) {
		s.logger.Info("Admin user exists")
		return nil
	}
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
