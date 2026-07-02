package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/getaudited/audited/internal/domain"
)

type UsersClickhouseRepository struct {
	db clickhousedb.Conn
}

func NewUsersClickhouseRepository(db clickhousedb.Conn) UsersClickhouseRepository {
	return UsersClickhouseRepository{
		db: db,
	}
}

func (r UsersClickhouseRepository) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT id, email, password, role, created_at FROM users WHERE email = ?`,
		email.String(),
	)
	return mapRowToUser(row)
}

func (r UsersClickhouseRepository) FindByID(ctx context.Context, id domain.ID) (*domain.User, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT id, email, password, role, created_at FROM users WHERE id = ?`,
		id.String(),
	)
	return mapRowToUser(row)
}

func (r UsersClickhouseRepository) Save(ctx context.Context, user *domain.User) error {
	exists, err := r.userExists(ctx, user.Email())
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrUserExists
	}

	err = r.db.Exec(
		ctx,
		`INSERT INTO users (id, email, password, role, created_at) VALUES (?, ?, ?, ?, ?)`,
		user.ID().String(),
		user.Email().String(),
		user.Password().String(),
		user.Role().String(),
		user.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("error saving user with email '%s' due to: %w", user.Email(), err)
	}

	return nil
}

func (r UsersClickhouseRepository) userExists(ctx context.Context, email domain.Email) (bool, error) {
	row := r.db.QueryRow(ctx, `SELECT id FROM users WHERE email = ?`, email.String())
	var id string
	err := row.Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error checking for user via email '%s': %w", email, err)
	}

	return true, nil
}
