package query

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type AllEventTypes struct {
	Action           *string
	PaginationParams PaginationParams
}

type eventTypesFinder interface {
	QueryAll(ctx context.Context, params AllEventTypes) (Pagination[*domain.EventType], error)
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
	return e.finder.QueryAll(ctx, q)
}
