package query

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type Pagination[T any] struct {
	Data        []T
	Total       int
	PerPage     int
	CurrentPage int
	TotalPages  int
}

type CursorPaginationResult[T any] struct {
	LastItemCursor string
	Data           []T
}

type CursorPaginationParams struct {
	Limit           *int
	StartFromCursor *string
}

type PaginationParams struct {
	Limit int
	Page  int
}

type eventTypeFinder interface {
	FindByAction(ctx context.Context, action string) (*domain.EventType, error)
}

type sourcesFinder interface {
	QueryAll(ctx context.Context, params PaginationParams) (Pagination[domain.Source], error)
}

type sourceByIDFinder interface {
	FindByID(ctx context.Context, id string) (*domain.Source, error)
}

type eventsFinder interface {
	QueryAll(ctx context.Context, params AllEventsParams, pagination CursorPaginationParams) (CursorPaginationResult[domain.Event], error)
}

type eventTypesFinder interface {
	QueryAll(ctx context.Context, params PaginationParams) (Pagination[*domain.EventType], error)
}
