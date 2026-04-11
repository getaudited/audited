package http

import (
	"context"
	"net/http"
	"time"

	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/friendsofgo/errors"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

const mockTenantID = "dummy-tenant-id"

func (h handlers) CreateEventType(c echo.Context) error {
	var body CreateEventTypeJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	var schema []byte
	if body.Schema != nil {
		schema = []byte(*body.Schema)
	}

	eventType := domain.EventType{
		Id:                           ulid.Make().String(),
		TenantID:                     mockTenantID,
		Version:                      1,
		Action:                       body.Action,
		TargetTypes:                  body.TargetTypes,
		ShouldValidateMetadataSchema: body.ShouldValidateMetadataSchema,
		Schema:                       schema,
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
	}

	err = h.application.Commands.CreateEventType.Execute(ctxFromEcho(c), command.CreateEventType{
		EventType: eventType,
	})
	if err != nil {
		return NewBadRequestError(err, "unable-to-create-event-type")
	}

	return c.JSON(http.StatusCreated, EventType{
		Id:                           eventType.Id,
		Action:                       eventType.Action,
		TargetTypes:                  eventType.TargetTypes,
		Schema:                       body.Schema,
		ShouldValidateMetadataSchema: eventType.ShouldValidateMetadataSchema,
		CreatedAt:                    eventType.CreatedAt,
		UpdatedAt:                    eventType.UpdatedAt,
	})
}

func (h handlers) GetEventTypeByID(c echo.Context, eventTypeAction EventTypeAction) error {
	et, err := h.application.Queries.EventTypeByAction.Execute(ctxFromEcho(c), query.EventTypeByAction{
		TenantID: mockTenantID,
		Action:   eventTypeAction,
	})
	if errors.Is(err, domain.ErrEventTypeNotFound) {
		return NewNotFoundError(err, "event-type-not-found")
	}
	if err != nil {
		return NewHandlerError(err, "error-querying-event-type")
	}

	schema := string(et.Schema)
	return c.JSON(http.StatusOK, EventType{
		Id:                           et.Id,
		Version:                      et.Version,
		Action:                       et.Action,
		TargetTypes:                  et.TargetTypes,
		Schema:                       &schema,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		CreatedAt:                    et.CreatedAt,
		UpdatedAt:                    et.UpdatedAt,
	})
}

func ctxFromEcho(c echo.Context) context.Context {
	return c.Request().Context()
}
