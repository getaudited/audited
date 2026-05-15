package query

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type AllTokens struct {
	SourceID domain.ID
}

type AllTokensHandler struct {
	repo domain.TokenRepository
}

func NewAllTokensHandler(repo domain.TokenRepository) AllTokensHandler {
	return AllTokensHandler{
		repo: repo,
	}
}

func (h AllTokensHandler) Execute(ctx context.Context, q AllTokens) ([]*domain.Token, error) {
	return h.repo.QueryAll(ctx, q.SourceID)
}
