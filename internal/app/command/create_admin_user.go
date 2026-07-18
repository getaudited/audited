package command

import (
	"context"
	"errors"

	"github.com/getaudited/audited/internal/domain"
)

type CreateAdminUser struct {
	Email    string
	Password string
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
	email, err := domain.NewEmail(cmd.Email)
	if err != nil {
		return err
	}
	
	password, err := domain.NewPassword(cmd.Password)
	if err != nil {
		return err
	}

	adminUser, err := c.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	if adminUser != nil {
		return domain.ErrUserExists
	}

	user, err := domain.NewUser(email, password, domain.UserRoleAdmin)
	if err != nil {
		return err
	}

	return c.repo.Save(ctx, user)
}
