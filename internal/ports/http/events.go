package http

import (
	"errors"
	"net/http"

	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

func (h handlers) CreateEvent(c echo.Context, params CreateEventParams) error {
	var body CreateEventJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	event, err := mapRequestToDomainEvent(body)
	if err != nil {
		return NewBadRequestError(err, "unable to create event")
	}

	err = h.application.Commands.CreateEvent.Execute(mapEchoCtxToCtx(c), command.CreateEvent{
		Event: event,
		Token: domain.TokenValue(params.XToken),
	})
	if errors.Is(err, domain.ErrTokenNotFound) {
		return NewHandlerErrorWithStatus(err, "token-not-found", http.StatusUnauthorized)
	}
	if errors.Is(err, domain.ErrEventTypeActionNotFound) {
		return NewHandlerErrorWithStatus(err, "event-type-action-not-found", http.StatusNotFound)
	}
	if err != nil {
		return NewHandlerError(err, "unable-to-create-event")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h handlers) GetEvents(c echo.Context, params GetEventsParams) error {
	var actorID domain.ID
	var targetID domain.ID

	if params.ActorId != nil {
		actorID = domain.ID(*params.ActorId)
	}

	if params.TargetId != nil {
		targetID = domain.ID(*params.TargetId)
	}

	result, err := h.application.Queries.Events.Execute(mapEchoCtxToCtx(c), query.AllEvents{
		SourceID:      domain.ID(params.SourceId),
		StartDate:     params.StartDate,
		EndDate:       params.EndDate,
		ActorID:       actorID,
		ActorName:     params.ActorName,
		TargetID:      targetID,
		Action:        params.Action,
		Limit:         params.Limit,
		StartingAfter: params.StartingAfter,
		EndingBefore:  params.EndingBefore,
	})
	if err != nil {
		return NewHandlerError(err, "error-querying-events")
	}

	return c.JSON(http.StatusOK, EventList{
		HasMore: result.HasMore,
		Data:    mapDomainEventsToEvents(result.Data),
	})
}

func (h handlers) GetEventByID(c echo.Context, eventID EventId) error {
	event, err := h.application.Queries.EventByID.Execute(mapEchoCtxToCtx(c), query.EventByID{
		ID: domain.ID(eventID),
	})
	if errors.Is(err, domain.ErrEventTypeExists) {
		return NewNotFoundError(err, "event-not-found")
	}
	if err != nil {
		return NewHandlerError(err, "unable-to-get-event-by-id")
	}

	return c.JSON(http.StatusOK, mapDomainEventToEvent(event))
}
