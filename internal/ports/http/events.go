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
	// TODO: check token
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
		Params: query.AllEventsParams{
			SourceID:  domain.ID(params.SourceId),
			StartDate: params.StartDate,
			EndDate:   params.EndDate,
			ActorID:   actorID,
			ActorName: params.ActorName,
			TargetID:  targetID,
		},
		CursorPaginationParams: query.CursorPaginationParams{
			Limit:           params.Limit,
			StartFromCursor: params.StartFrom,
		},
	})
	if err != nil {
		return NewHandlerError(err, "error-querying-events")
	}

	return c.JSON(http.StatusOK, EventList{
		LastItemCursor: result.LastItemCursor,
		Data:           mapDomainEventsToEvents(result.Data),
	})
}

func mapDomainEventsToEvents(events []domain.Event) []Event {
	r := make([]Event, len(events))

	for i, e := range events {
		r[i] = mapDomainEventToEvent(e)
	}

	return r
}

func mapDomainEventToEvent(e domain.Event) Event {
	targets := make([]Target, len(e.Targets()))

	for i, t := range e.Targets() {
		targets[i] = Target{
			Id:         t.ID,
			Name:       t.Name,
			Metadata:   t.Metadata,
			TargetType: t.TargetType,
		}
	}

	return Event{
		Id:       e.ID().String(),
		Metadata: e.Metadata(),
		SourceId: e.SourceID().String(),
		Actor: Actor{
			Id:        e.Actor().ID,
			Name:      e.Actor().Name,
			ActorType: e.Actor().ActorType,
			Metadata:  e.Actor().Metadata,
		},
		Context: Context{
			Location:  e.Context().Location,
			UserAgent: e.Context().UserAgent,
		},
		Targets:    targets,
		Version:    e.Version(),
		OccurredAt: e.OccurredAt().String(),
	}
}
