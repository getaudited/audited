package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"
	"github.com/tachyonhqdev/webhooks/internal/common/logs"
	"github.com/tachyonhqdev/webhooks/tests/emulators"
)

type GenericResponse struct {
	Message string `json:"message"`
}

type webhookCalls struct {
	mu   *sync.RWMutex
	data map[string]emulators.WebhookCall
}

func main() {
	logger := logs.New()
	router := echo.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	wbCalls := webhookCalls{
		mu:   &sync.RWMutex{},
		data: make(map[string]emulators.WebhookCall),
	}

	webhooks := router.Group("/webhooks")
	webhooks.POST("/*", func(c echo.Context) error {
		wbCalls.mu.Lock()
		defer wbCalls.mu.Unlock()

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			logger.Error(err.Error())
			return c.JSON(http.StatusInternalServerError, GenericResponse{
				Message: "error reading body",
			})
		}

		wh, err := standardwebhooks.NewWebhook(emulators.TestEndpointSecretBase64)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, GenericResponse{
				Message: fmt.Sprintf("error creating standard webhook: %s", err),
			})
		}

		err = wh.Verify(body, c.Request().Header)
		if err != nil {
			return c.JSON(http.StatusForbidden, GenericResponse{
				Message: fmt.Sprintf("error verifying signature: %s", err.Error())},
			)
		}

		headers := make(map[string]string)
		for name, value := range c.Request().Header {
			headers[name] = value[0]
		}

		wbCalls.data[c.Request().URL.String()] = emulators.WebhookCall{
			Headers: headers,
			Body:    string(body),
		}

		return c.JSON(http.StatusOK, GenericResponse{Message: "success"})
	})

	webhooks.GET("/calls", func(c echo.Context) error {
		wbCalls.mu.RLock()
		defer wbCalls.mu.RUnlock()

		return c.JSON(http.StatusOK, emulators.GetWebhookCallsResponse{Calls: wbCalls.data})
	})

	err := router.Start(":9090")
	if err != nil {
		logger.Error("emulator exited with an error", "error", err)
		os.Exit(1)
	}
}
