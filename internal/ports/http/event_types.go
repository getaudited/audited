package http

import (
	"net/http"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

func (h handlers) GetEventTypes(c echo.Context, params GetEventTypesParams) error {
	result, err := h.application.Queries.AllEventTypes.Execute(mapEchoCtxToCtx(c), query.AllEventTypes{
		Action:           params.Action,
		PaginationParams: mapToQueryPaginationParams(params.Page, params.Limit),
	})
	if err != nil {
		return NewHandlerError(err, "unable-to-retrieve-event-types")
	}

	return c.JSON(http.StatusOK, mapToEventTypeList(result))
}

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

	et := domain.EventType{
		Action:                       body.Action,
		ShouldValidateMetadataSchema: body.ShouldValidateMetadataSchema,
		LastVersion:                  domain.NewEventTypeVersion(body.TargetTypes, schema),
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
	}

	err = h.application.Commands.CreateEventType.Execute(mapEchoCtxToCtx(c), command.CreateEventType{
		EventType: et,
	})
	if errors.Is(err, domain.ErrEventTypeExists) {
		return NewHandlerErrorWithStatus(err, "error-event-type-exists", http.StatusConflict)
	}
	if err != nil {
		return NewBadRequestError(err, "unable-to-create-event-type")
	}

	return c.JSON(http.StatusCreated, EventType{
		Action:                       et.Action,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		Version:                      et.LastVersion.Version,
		Schema:                       nil,
		TargetTypes:                  et.LastVersion.TargetTypes,
		CreatedAt:                    et.CreatedAt,
	})
}

func (h handlers) GetEventTypeByID(c echo.Context, action Action) error {
	et, err := h.application.Queries.EventTypeByAction.Execute(mapEchoCtxToCtx(c), query.EventTypeByAction{
		Action: action,
	})
	if errors.Is(err, domain.ErrEventTypeNotFound) {
		return NewNotFoundError(err, "event-type-not-found")
	}
	if err != nil {
		return NewHandlerError(err, "error-querying-event-type")
	}

	return c.JSON(http.StatusOK, mapToEventType(et))
}

func (h handlers) DeleteEventType(c echo.Context, action Action) error {
	err := h.application.Commands.DeleteEventType.Execute(mapEchoCtxToCtx(c), command.DeleteEventType{
		Action: action,
	})
	if err != nil {
		return NewHandlerError(err, "error-deleting-event-type")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h handlers) CreateEventTypeVersion(c echo.Context, action Action) error {
	var body CreateEventTypeVersionJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	var schema domain.Schema
	if body.Schema != nil {
		schema = domain.Schema(*body.Schema)
	}

	err = h.application.Commands.CreateEventTypeVersion.Execute(mapEchoCtxToCtx(c), command.CreateEventTypeVersion{
		Schema:      schema,
		Action:      action,
		TargetTypes: body.TargetTypes,
	})
	if errors.Is(err, domain.ErrEventTypeNotFound) {
		return NewNotFoundError(err, "event-type-not-found")
	}
	if err != nil {
		return NewHandlerError(err, "unable-to-create-event-type-version")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h handlers) RollbackEventTypeVersion(c echo.Context, action Action) error {
	err := h.application.Commands.RollbackEventTypeVersion.Execute(mapEchoCtxToCtx(c), command.RollbackEventTypeVersion{
		Action: action,
	})
	if errors.Is(err, domain.ErrVersionOneOfEventTypeCannotBeRolledBack) {
		return NewHandlerErrorWithStatus(err, "unable-to-rollback-version-one-of-event-type", http.StatusConflict)
	}
	if err != nil {
		return NewHandlerError(err, "unable-to-rollback-event-event-type")
	}

	return c.NoContent(http.StatusNoContent)
}
