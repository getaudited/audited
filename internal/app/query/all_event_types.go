package query

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type AllEventTypes struct {
	PaginationParams PaginationParams
}

type AllEventTypesHandler struct {
	finder eventTypesFinder
}

func NewAllEventTypesHandler(finder eventTypesFinder) AllEventTypesHandler {
	return AllEventTypesHandler{
		finder: finder,
	}
}

func (e AllEventTypesHandler) Execute(ctx context.Context, q AllEventTypes) (Pagination[*domain.EventType], error) {
	return e.finder.QueryAll(ctx, q.PaginationParams)
}
