package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type RollbackEventTypeVersion struct {
	Action string
}

type RollbackEventTypeVersionHandler struct {
	repo domain.EventTypeRepository
}

func NewRollbackEventTypeVersionHandler(repo domain.EventTypeRepository) RollbackEventTypeVersionHandler {
	return RollbackEventTypeVersionHandler{
		repo: repo,
	}
}

func (c RollbackEventTypeVersionHandler) Execute(ctx context.Context, cmd RollbackEventTypeVersion) error {
	return c.repo.RollbackVersion(ctx, cmd.Action)
}
