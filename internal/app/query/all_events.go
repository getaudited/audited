package query

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type AllEvents struct {
	SourceID               domain.ID
	CursorPaginationParams CursorPaginationParams
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
	return e.finder.QueryAll(ctx, q.SourceID, q.CursorPaginationParams)
}
