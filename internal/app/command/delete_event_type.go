package command

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type DeleteEventType struct {
	Action string
}

type DeleteEventTypeHandler struct {
	repo domain.EventTypeRepository
}

func NewDeleteEventTypeHandler(repo domain.EventTypeRepository) DeleteEventTypeHandler {
	return DeleteEventTypeHandler{
		repo: repo,
	}
}

func (c DeleteEventTypeHandler) Execute(ctx context.Context, cmd DeleteEventType) error {
	return c.repo.Delete(ctx, cmd.Action)
}
