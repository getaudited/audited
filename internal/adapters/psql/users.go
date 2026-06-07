package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/domain"
)

type UserPsqlRepository struct {
	db *sql.DB
}

func NewUsersPsqlRepository(db *sql.DB) UserPsqlRepository {
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
