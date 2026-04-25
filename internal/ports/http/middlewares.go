package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/firminochangani/audited/internal/common/logs"
	"github.com/friendsofgo/errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func loggerMiddleware(logger *logs.Logger, isDebug bool) echo.MiddlewareFunc {
	if isDebug {
		return middleware.RequestLogger()
	}

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogError:     false,
		LogRequestID: true,
		LogHeaders:   []string{echo.HeaderXCorrelationID},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			lvl := slog.LevelInfo
			msg := "request handled successfully"
			if v.Error != nil || v.Status > http.StatusBadRequest {
				lvl = slog.LevelError
				msg = "request handled with an error"
			}

			logger.LogAttrs(
				c.Request().Context(),
				lvl,
				msg,
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.String("correlation_id", v.RequestID),
			)

			return nil
		},
	})
}

func bearerAuthMiddleware(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
	// implement me
	return nil
}

func tokenAuthMiddleware(_ context.Context, input *openapi3filter.AuthenticationInput) error {
	token := input.RequestValidationInput.Request.Header.Get("x-token")
	if token == strings.TrimSpace(token) {
		return errors.New("missing token")
	}

	// implement me
	return nil
}

func errorHandler(logger *logs.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		logger.Debug("Handler failed", "error", err)

		code := http.StatusInternalServerError
		errorSlug := "internal-server-error"
		errorMessage := err.Error()

		switch e := err.(type) {
		case *echo.HTTPError:
			code = e.Code
			errorMessage = e.Error()
		case *HandlerError:
			code = e.Code
			errorSlug = e.Slug()
			errorMessage = e.Error()
		}

		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, ErrorSchema{
				Error:   errorSlug,
				Message: errorMessage,
			})
		}
		if err != nil {
			logger.Debug("Failed to send error response", "error", err)
		}
	}
}
