package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/tachyonhqdev/webhooks/internal/app/command"
	"github.com/tachyonhqdev/webhooks/internal/domain"
)

func (h handlers) CreateEvent(c echo.Context) error {
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

	err = h.application.Commands.CreateEvent.Execute(ctxFromEcho(c), command.CreateEvent{
		Event: domain.Event{
			Id:         ulid.Make().String(),
			OccurredAt: body.OccurredAt,
			Version:    body.Version,
			// TODO: add tenant_id
			Actor: domain.Actor{
				Id:        body.Actor.Id,
				ActorType: body.Actor.Type,
				Name:      body.Actor.Name,
				Metadata:  body.Actor.Metadata,
			},
			Targets: targets,
			Context: domain.Context{
				Location:  body.Context.Location,
				UserAgent: body.Context.UserAgent,
			},
			Metadata: body.Metadata,
		},
	})
	if err != nil {
		return NewHandlerError(err, "unable-to-create-event")
	}

	return c.NoContent(http.StatusNoContent)
}
