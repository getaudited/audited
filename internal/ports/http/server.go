package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	libhttp "net/http"
	"strings"
	"time"

	"github.com/firminochangani/audited/internal/app"
	"github.com/firminochangani/audited/internal/common/logs"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"
)

type Server struct {
	logger *logs.Logger
	s      *libhttp.Server
}

type Config struct {
	Application        *app.App
	Port               int
	AllowedCorsOrigin  []string
	Logger             *logs.Logger
	IsDebug            bool
	Ctx                context.Context
	JwtSecret          string
	WebFrontendEnabled bool
}

func NewServer(config Config) (*Server, error) {
	router := echo.New()

	spec, err := GetSwagger()
	if err != nil {
		return nil, err
	}

	routerHandlers := &handlers{
		application: config.Application,
		jwtSecret:   config.JwtSecret,
	}

	registerMiddlewares(router, spec, config)
	registerAdditionalHandlers(router, routerHandlers, spec)
	RegisterHandlers(router, routerHandlers)

	return &Server{
		logger: config.Logger,
		s: &libhttp.Server{
			Handler:           router,
			ReadTimeout:       time.Second * 30,
			ReadHeaderTimeout: time.Second * 30,
			WriteTimeout:      time.Second * 30,
			IdleTimeout:       time.Second * 30,
			Addr:              fmt.Sprintf(":%d", config.Port),
			BaseContext: func(listener net.Listener) context.Context {
				return config.Ctx
			},
		},
	}, nil
}

func (s *Server) Start() error {
	s.logger.Info("http server is running", "port", s.s.Addr)
	err := s.s.ListenAndServe()
	if err != nil && !errors.Is(err, libhttp.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down the http server")
	return s.s.Shutdown(ctx)
}

func registerMiddlewares(router *echo.Echo, spec *openapi3.T, config Config) {
	router.HTTPErrorHandler = errorHandler(config.Logger)
	router.Use(middleware.RequestID())
	router.Use(middleware.Recover())
	router.Use(loggerMiddleware(config.Logger, config.IsDebug))
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.AllowedCorsOrigin,
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	spec.Servers = nil
	router.Use(oapimiddleware.OapiRequestValidatorWithOptions(
		spec,
		&oapimiddleware.Options{
			ErrorHandler: nil,
			Options: openapi3filter.Options{
				AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
					return nil
				},
			},
			ParamDecoder: nil,
			UserData:     nil,
			Skipper: func(c echo.Context) bool {
				return strings.HasPrefix(c.Request().URL.Path, "/openapi.json") || strings.HasPrefix(c.Request().URL.Path, "/health")
			},
			MultiErrorHandler:     nil,
			SilenceServersWarning: false,
		},
	))
}
func registerAdditionalHandlers(router *echo.Echo, h *handlers, spec *openapi3.T) {
	router.GET("/health", h.HealthCheck)
	router.GET("/openapi.json", func(c echo.Context) error {
		return c.JSON(libhttp.StatusOK, spec)
	})
}
