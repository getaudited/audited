package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

func (h handlers) CreateSource(c echo.Context) error {
	var body CreateSourceJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	source, err := domain.NewSource(body.Name)
	if err != nil {
		return err
	}

	err = h.application.Commands.CreateSource.Execute(ctxFromEcho(c), command.CreateSource{
		Source: source,
	})
	if err != nil {
		return NewHandlerError(err, "error-creating-source")
	}

	return c.JSON(http.StatusCreated, Source{
		Id:        source.ID().String(),
		Name:      source.Name(),
		CreatedAt: source.CreatedAt(),
		UpdatedAt: source.UpdatedAt(),
	})
}
