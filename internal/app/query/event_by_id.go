package query

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type EventByID struct {
	ID domain.ID
}

type EventByIDHandler struct {
	repo domain.EventRepository
}

func NewEventByIDHandler(repo domain.EventRepository) EventByIDHandler {
	return EventByIDHandler{
		repo: repo,
	}
}

func (h EventByIDHandler) Execute(ctx context.Context, q EventByID) (domain.Event, error) {
	return h.repo.FindByID(ctx, q.ID)
}
