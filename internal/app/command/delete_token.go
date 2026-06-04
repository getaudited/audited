package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type DeleteToken struct {
	TokenID  domain.ID
	SourceID domain.ID
}

type DeleteTokenHandler struct {
	repo domain.TokenRepository
}

func NewDeleteTokenHandler(repo domain.TokenRepository) DeleteTokenHandler {
	return DeleteTokenHandler{
		repo: repo,
	}
}

func (c DeleteTokenHandler) Execute(ctx context.Context, cmd DeleteToken) error {
	return c.repo.Delete(ctx, cmd.TokenID, cmd.SourceID)
}
