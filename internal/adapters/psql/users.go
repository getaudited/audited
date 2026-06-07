package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/domain"
)

type UserPsqlRepository struct {
	db boil.ContextExecutor
}

func NewUsersPsqlRepository(db boil.ContextExecutor) UserPsqlRepository {
	return UserPsqlRepository{
		db: db,
	}
}

func (u UserPsqlRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	row, err := models.Users(models.UserWhere.Email.EQ(email.String())).One(ctx, u.db)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying user via email '%s': %w", email, err)
	}

	return domain.MarshallToUser(row.ID, row.Email, row.Password, row.Role, row.CreatedAt), nil
}

func (u UserPsqlRepository) Save(ctx context.Context, user *domain.User) error {
	row := models.User{
		ID:        user.ID().String(),
		Email:     user.Email().String(),
		Password:  user.Password().String(),
		Role:      user.Role().String(),
		CreatedAt: user.CreatedAt(),
	}
	err := row.Upsert(ctx, u.db, false, []string{"email"}, boil.Infer(), boil.Infer())
	if err != nil {
		return fmt.Errorf("error saving user: %w", err)
	}

	return nil
}
