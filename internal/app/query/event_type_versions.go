package query

import (
	"context"
)

type EventTypeVersions struct {
	Action string
}

type EventTypeVersionsHandler struct {
	finder eventTypeFinder
}

func NewEventTypeVersionsHandler(finder eventTypeFinder) EventTypeVersionsHandler {
	return EventTypeVersionsHandler{
		finder: finder,
	}
}

func (e EventTypeVersionsHandler) Execute(ctx context.Context, q EventTypeVersions) ([]EventType, error) {
	return e.finder.AllVersionsByAction(ctx, q.Action)
}
