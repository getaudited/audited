package command

import (
	"context"

	"github.com/tachyonhqdev/webhooks/internal/domain"
)

type CreateEvent struct {
	Event domain.Event
}

type CreateEventHandler struct {
	repo domain.EventRepository
}

func NewCreateEventHandler(repo domain.EventRepository) CreateEventHandler {
	return CreateEventHandler{
		repo: repo,
	}
}

func (c CreateEventHandler) Execute(ctx context.Context, cmd CreateEvent) error {
	return c.repo.Save(ctx, cmd.Event)
}
