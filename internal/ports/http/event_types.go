package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/tachyonhqdev/webhooks/internal/app/command"
	"github.com/tachyonhqdev/webhooks/internal/domain"
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
		TenantID:                     "dummy-tenant-id",
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

func ctxFromEcho(c echo.Context) context.Context {
	return c.Request().Context()
}
