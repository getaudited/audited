package query

import (
	"context"
	"time"

	"github.com/getaudited/audited/internal/domain"
)

type AllEvents struct {
	SourceID      domain.ID
	StartDate     *time.Time
	EndDate       *time.Time
	ActorID       domain.ID
	ActorName     *string
	TargetID      domain.ID
	Action        *string
	Limit         *int
	StartingAfter *string
	EndingBefore  *string
}

type AllEventsHandler struct {
	finder eventsFinder
}

func NewAllEventsHandler(finder eventsFinder) AllEventsHandler {
	return AllEventsHandler{
		finder: finder,
	}
}

func (e AllEventsHandler) Execute(ctx context.Context, q AllEvents) (CursorPaginationResult[domain.Event], error) {
	return e.finder.QueryAll(ctx, q)
}
