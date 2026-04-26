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
	Cursor Cursor
	Data   []T
}

type CursorPaginationParams struct {
	Limit  *int
	Cursor *string
}

type Cursor struct {
	Next string
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

type eventsFinder interface {
	QueryAll(ctx context.Context, params CursorPaginationParams) (CursorPaginationResult[domain.Event], error)
}
