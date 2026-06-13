package query

import (
	"context"
)

type AllEventTypes struct {
	Action           *string
	PaginationParams PaginationParams
}

type AllEventTypesHandler struct {
	finder eventTypeFinder
}

func NewAllEventTypesHandler(finder eventTypeFinder) AllEventTypesHandler {
	return AllEventTypesHandler{
		finder: finder,
	}
}

func (e AllEventTypesHandler) Execute(ctx context.Context, q AllEventTypes) (Pagination[EventType], error) {
	return e.finder.QueryAll(ctx, q)
}
