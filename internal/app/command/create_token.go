package command

import (
	"context"

	"github.com/firminochangani/audited/internal/domain"
)

type CreateToken struct {
	Token domain.Token
}

type CreateTokenHandler struct {
	repo domain.TokenRepository
}

func NewCreateTokenHandler(repo domain.TokenRepository) CreateTokenHandler {
	return CreateTokenHandler{
		repo: repo,
	}
}

func (c CreateTokenHandler) Execute(ctx context.Context, cmd CreateToken) error {
	return c.repo.Save(ctx, cmd.Token)
}
