package command

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type CreateSource struct {
	Source *domain.Source
}

type CreateSourceHandler struct {
	repo domain.SourceRepository
}

func NewCreateSourceHandler(repo domain.SourceRepository) CreateSourceHandler {
	return CreateSourceHandler{
		repo: repo,
	}
}

func (c CreateSourceHandler) Execute(ctx context.Context, cmd CreateSource) error {
	return c.repo.Save(ctx, cmd.Source)
}
