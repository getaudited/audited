package query

import (
	"context"
	"time"

	"github.com/getaudited/audited/internal/domain"
)

type Pagination[T any] struct {
	Data        []T
	Total       int
	PerPage     int
	CurrentPage int
	TotalPages  int
}

type CursorPaginationResult[T any] struct {
	HasMore bool
	Data    []T
}

type CursorPaginationParams struct {
	Limit           *int
	StartFromCursor *string
}

type PaginationParams struct {
	Limit int
	Page  int
}

type EventType struct {
	Action                       string
	ShouldValidateMetadataSchema bool
	Version                      int
	TargetTypes                  []string
	Schema                       string
	CreatedAt                    time.Time
}

type eventTypeFinder interface {
	FindByAction(ctx context.Context, action string) (EventType, error)
	AllVersionsByAction(ctx context.Context, action string) ([]EventType, error)
	QueryAll(ctx context.Context, params AllEventTypes) (Pagination[EventType], error)
}

type sourceByIDFinder interface {
	FindByID(ctx context.Context, id string) (*domain.Source, error)
}

type eventsFinder interface {
	QueryAll(ctx context.Context, params AllEvents) (CursorPaginationResult[domain.Event], error)
}
