package query

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type Pagination[T any] struct {
	Data []T
}

type PaginationParams struct {
	Limit int
	Page  int
}

type eventTypeFinder interface {
	FindByAction(ctx context.Context, tenantID string, action string) (*domain.EventType, error)
}
