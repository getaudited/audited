package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type CreateEventType struct {
	EventType domain.EventType
}

type CreateEventTypeHandler struct {
	repo domain.EventTypeRepository
}

func NewCreateEventTypeHandler(repo domain.EventTypeRepository) CreateEventTypeHandler {
	return CreateEventTypeHandler{
		repo: repo,
	}
}

func (c CreateEventTypeHandler) Execute(ctx context.Context, cmd CreateEventType) error {
	return c.repo.Save(ctx, cmd.EventType)
}
