package http

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"

	"github.com/getaudited/audited/internal/common/logs"
)

const (
	ctxUserID = "user_id"
)

var (
	errMissingToken = errors.New("missing x-token")
)

func loggerMiddleware(logger *logs.Logger) echo.MiddlewareFunc {
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

type JWTMiddleware struct {
	publicKey *ecdsa.PublicKey
}

func NewJWTMiddleware(publicKey *ecdsa.PublicKey) *JWTMiddleware {
	return &JWTMiddleware{
		publicKey: publicKey,
	}
}

func (m *JWTMiddleware) Authenticate(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	authToken := input.RequestValidationInput.Request.Header.Get(echo.HeaderAuthorization)
	if strings.TrimSpace(authToken) == "" {
		return errors.New("missing token")
	}

	token, err := jwt.ParseWithClaims(strings.TrimPrefix(authToken, "Bearer "), &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.New("incorrect signing method")
		}

		return m.publicKey, nil
	})
	if err != nil {
		return fmt.Errorf("error parsing the JWT provided: %w", err)
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("error parsing the JWT provided: %w", err)
	}

	eCtx := oapimiddleware.GetEchoContext(ctx)
	eCtx.Set(ctxUserID, claims.Subject)

	return nil
}

func tokenAuthMiddleware(_ context.Context, input *openapi3filter.AuthenticationInput) error {
	token := input.RequestValidationInput.Request.Header.Get("x-token")
	if strings.TrimSpace(token) == "" {
		return errMissingToken
	}

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

func mapRetrieveUserIdFromCtx(c echo.Context) (string, error) {
	userID, ok := c.Get(ctxUserID).(string)
	if !ok {
		return "", errors.New("unable to retrieve user_id from request's context")
	}

	return userID, nil
}
