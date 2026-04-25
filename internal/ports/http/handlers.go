package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app"
	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

var _ ServerInterface = (*handlers)(nil)

type handlers struct {
	application *app.App
}

func (h handlers) ArchiveEventType(c echo.Context, eventId EventId) error {
	return nil
}

func (h handlers) GetEventTypes(c echo.Context) error {
	return nil
}

func (h handlers) DeleteToken(c echo.Context, sourceId SourceId, tokenId TokenId) error {
	return nil
}

func (h handlers) CreateToken(c echo.Context, sourceId SourceId) error {
	var body CreateSourceJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	token, err := domain.NewToken(domain.ID(sourceId), body.Name)
	if err != nil {
		return NewBadRequestError(err, "error-validating-data")
	}

	err = h.application.Commands.CreateToken.Execute(ctxFromEcho(c), command.CreateToken{
		Token: *token,
	})
	if err != nil {
		return NewHandlerError(err, "error-creating-token")
	}

	return c.JSON(http.StatusCreated, Token{
		Name:      token.Name(),
		SourceId:  token.SourceID().String(),
		Value:     token.ID().String(),
		CreatedAt: token.CreatedAt(),
	})
}

func (h handlers) GetEvents(c echo.Context, params GetEventsParams) error {
	//TODO implement me
	return nil
}

func (h handlers) GetSourceByID(c echo.Context, sourceId SourceId) error {
	//TODO implement me
	return nil
}

func (h handlers) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
