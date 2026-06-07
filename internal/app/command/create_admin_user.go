package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type CreateAdminUser struct {
	User *domain.User
}

type CreateAdminUserHandler struct {
	repo domain.UserRepository
}

func NewCreateAdminUserHandler(repo domain.UserRepository) CreateAdminUserHandler {
	return CreateAdminUserHandler{
		repo: repo,
	}
}

func (c CreateAdminUserHandler) Execute(ctx context.Context, cmd CreateAdminUser) error {
	return c.repo.Save(ctx, cmd.User)
}
