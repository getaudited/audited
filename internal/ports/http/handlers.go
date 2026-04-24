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
	//TODO implement me
	panic("implement me")
}

func (h handlers) GetEventTypes(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) DeleteToken(ctx echo.Context, sourceId SourceId, tokenId TokenId) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) CreateSourceToken(ctx echo.Context, sourceId SourceId) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) GetEvents(ctx echo.Context, params GetEventsParams) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) GetSources(ctx echo.Context, params GetSourcesParams) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) GetSourceByID(ctx echo.Context, sourceId SourceId) error {
	//TODO implement me
	panic("implement me")
}

func (h handlers) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
