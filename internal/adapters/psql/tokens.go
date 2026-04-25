package psql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/firminochangani/audited/internal/adapters/models"
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

func (t TokensPsqlRepository) Delete(ctx context.Context, id, sourceID domain.ID) error {
	row := models.Token{
		ID:       id.String(),
		SourceID: sourceID.String(),
	}
	_, err := row.Delete(ctx, t.db)
	if err != nil {
		return fmt.Errorf("error deleting token with id '%s': %v", id, err)
	}

	return nil
}
