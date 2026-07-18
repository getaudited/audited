package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type CreateEventTypeVersion struct {
	Action      string
	TargetTypes []string
	Schema      domain.Schema
}

type CreateEventTypeVersionHandler struct {
	repo domain.EventTypeRepository
}

func NewCreateEventTypeVersionHandler(repo domain.EventTypeRepository) CreateEventTypeVersionHandler {
	return CreateEventTypeVersionHandler{
		repo: repo,
	}
}

func (c CreateEventTypeVersionHandler) Execute(ctx context.Context, cmd CreateEventTypeVersion) error {
	return c.repo.SaveVersion(ctx, cmd.Action, cmd.TargetTypes, cmd.Schema)
}
