package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app"
	"github.com/labstack/echo/v4"
)

var _ ServerInterface = (*handlers)(nil)

type handlers struct {
	application *app.App
	jwtSecret   string
}

func (h handlers) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
