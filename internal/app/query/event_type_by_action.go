package query

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
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

func (e EventTypeByActionHandler) Execute(ctx context.Context, q EventTypeByAction) (*domain.EventType, error) {
	return e.finder.FindByAction(ctx, q.Action)
}
