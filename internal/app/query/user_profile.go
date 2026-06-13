package query

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type UserProfile struct {
	UserID domain.ID
}

type UserProfileHandler struct {
	repo domain.UserRepository
}

func NewUserProfileHandler(repo domain.UserRepository) UserProfileHandler {
	return UserProfileHandler{
		repo: repo,
	}
}

func (h UserProfileHandler) Execute(ctx context.Context, q UserProfile) (*domain.User, error) {
	return h.repo.FindByID(ctx, q.UserID)
}
