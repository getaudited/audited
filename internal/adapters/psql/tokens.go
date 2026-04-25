package psql

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/firminochangani/audited/internal/domain"
)

type TokensPsqlRepository struct {
	db *sql.DB
}

func NewTokensPsqlRepository(db *sql.DB) *TokensPsqlRepository {
	return &TokensPsqlRepository{db: db}
}

func (t TokensPsqlRepository) Save(ctx context.Context, token domain.Token) error {
	row := mapDomainTokenToModelToken(token)
	return row.Insert(ctx, t.db, boil.Infer())
}
