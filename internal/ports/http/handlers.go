package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app"
	"github.com/labstack/echo/v4"
)

var _ ServerInterface = (*handlers)(nil)

type handlers struct {
	application *app.App
}

func (h handlers) ArchiveEventType(ctx echo.Context, eventId EventId) error {
	return nil
}

func (h handlers) GetEventTypes(ctx echo.Context) error {
	return nil
}

func (h handlers) DeleteToken(ctx echo.Context, sourceId SourceId, tokenId TokenId) error {
	return nil
}

func (h handlers) CreateSourceToken(ctx echo.Context, sourceId SourceId) error {
	return nil
}

func (h handlers) GetEvents(ctx echo.Context, params GetEventsParams) error {
	//TODO implement me
	return nil
}

func (h handlers) GetSources(ctx echo.Context, params GetSourcesParams) error {
	//TODO implement me
	return nil
}

func (h handlers) GetSourceByID(ctx echo.Context, sourceId SourceId) error {
	//TODO implement me
	return nil
}

func (h handlers) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
