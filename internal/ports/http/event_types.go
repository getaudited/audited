package http

import (
	"net/http"
	"time"

	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/friendsofgo/errors"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

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
		Version:                      1,
		Action:                       body.Action,
		TargetTypes:                  body.TargetTypes,
		ShouldValidateMetadataSchema: body.ShouldValidateMetadataSchema,
		Schema:                       schema,
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
	}

	err = h.application.Commands.CreateEventType.Execute(mapEchoCtxToCtx(c), command.CreateEventType{
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

func (h handlers) GetEventTypeByID(c echo.Context, action EventTypeAction) error {
	et, err := h.application.Queries.EventTypeByAction.Execute(mapEchoCtxToCtx(c), query.EventTypeByName{
		Action: action,
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

func (h handlers) DeleteEventType(c echo.Context, action EventTypeAction) error {
	err := h.application.Commands.DeleteEventType.Execute(mapEchoCtxToCtx(c), command.DeleteEventType{
		Action: action,
	})
	if err != nil {
		return NewHandlerError(err, "error-deleting-event-type")
	}

	return c.NoContent(http.StatusNoContent)
}
