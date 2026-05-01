package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/labstack/echo/v4"
)

func (h handlers) CreateEvent(c echo.Context, params CreateEventParams) error {
	// TODO: check token
	var body CreateEventJSONBody
	err := c.Bind(&body)
	if err != nil {
		return NewBadRequestError(err, "unable-to-parse-body")
	}

	targets := make([]domain.Target, len(body.Targets))
	for i, target := range body.Targets {
		targets[i] = domain.Target{
			ID:         target.Id,
			Name:       target.Name,
			TargetType: target.Type,
			Metadata:   target.Metadata,
		}
	}

	event, err := domain.NewEvent(
		domain.ID(body.SourceId),
		body.Version,
		domain.Actor{
			ID:        body.Actor.Id,
			ActorType: body.Actor.Type,
			Name:      body.Actor.Name,
			Metadata:  body.Actor.Metadata,
		},
		targets,
		domain.Context{
			Location:  body.Context.Location,
			UserAgent: body.Context.UserAgent,
		},
		body.Metadata,
		body.OccurredAt,
	)
	if err != nil {
		return NewBadRequestError(err, "unable to create event")
	}

	err = h.application.Commands.CreateEvent.Execute(ctxFromEcho(c), command.CreateEvent{
		Event: event,
	})
	if err != nil {
		return NewHandlerError(err, "unable-to-create-event")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h handlers) GetEvents(c echo.Context, params GetEventsParams) error {
	result, err := h.application.Queries.Events.Execute(ctxFromEcho(c), query.AllEvents{
		SourceID: domain.ID(params.SourceId),
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
