package main

import (
	"cmp"
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
	"github.com/getaudited/audited/internal/domain"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"

	"github.com/getaudited/audited/internal/app"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/common/logs"
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

	jwtPrivateKey, err := s.parseJwtPrivateKey()
	if err != nil {
		return err
	}

	application, closer, err := resolveAppFromDatabase(ctx, logger, jwtPrivateKey, config)
	if err != nil {
		return err
	}

	// Set up Admin user
	err = s.createAdminUserIfNotExists(ctx, application)
	if err != nil {
		return err
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

		_ = closer.Close()

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

func resolveAppFromDatabase(ctx context.Context, logger *logs.Logger, jwtPrivateKey *ecdsa.PrivateKey, config *Config) (*app.App, Closer, error) {
	if strings.TrimSpace(config.DatabaseURL) != "" {
		return NewPostgresApp(ctx, logger, jwtPrivateKey, config)
	}

	if strings.TrimSpace(config.ClickhouseDatabaseURL) != "" {
		return NewClickhouseApp(ctx, config, jwtPrivateKey)
	}

	return nil, nil, errors.New("no database has been set")
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
