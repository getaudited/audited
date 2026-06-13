package query

import (
	"context"
)

type EventTypeByAction struct {
	Action string
}

type EventTypeByActionHandler struct {
	finder eventTypeFinder
}

func NewEventTypeByActionHandler(finder eventTypeFinder) EventTypeByActionHandler {
	return EventTypeByActionHandler{
		finder: finder,
	}
}

func (e EventTypeByActionHandler) Execute(ctx context.Context, q EventTypeByAction) (EventType, error) {
	return e.finder.FindByAction(ctx, q.Action)
}
