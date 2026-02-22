package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tachyonhqdev/webhooks/internal/app"
)

var _ ServerInterface = (*handlers)(nil)

type handlers struct {
	application *app.App
	jwtSecret   string
}

func (h handlers) CreateEvent(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
