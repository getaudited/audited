package http

import (
	"net/http"

	"github.com/firminochangani/audited/internal/app/command"
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
			Id:         target.Id,
			Name:       target.Name,
			TargetType: target.Type,
			Metadata:   target.Metadata,
		}
	}

	event, err := domain.NewEvent(
		domain.ID(body.SourceId),
		body.Version,
		domain.Actor{
			Id:        body.Actor.Id,
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
