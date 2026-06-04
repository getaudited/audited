package psql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/domain"
)

type TokensPsqlRepository struct {
	db *sql.DB
}

func NewTokensPsqlRepository(db *sql.DB) *TokensPsqlRepository {
	return &TokensPsqlRepository{db: db}
}

func (r TokensPsqlRepository) Save(ctx context.Context, token *domain.Token) error {
	row := mapDomainTokenToModelToken(token)
	return row.Insert(ctx, r.db, boil.Infer())
}

func (r TokensPsqlRepository) Delete(ctx context.Context, id, sourceID domain.ID) error {
	row := models.Token{
		ID:       id.String(),
		SourceID: sourceID.String(),
	}
	_, err := row.Delete(ctx, r.db)
	if err != nil {
		return fmt.Errorf("error deleting token with id '%s': %v", id, err)
	}

	return nil
}

func (r TokensPsqlRepository) QueryAll(ctx context.Context, sourceID domain.ID) ([]*domain.Token, error) {
	rows, err := models.Tokens(models.TokenWhere.SourceID.EQ(sourceID.String())).All(ctx, r.db)
	if err != nil {
		return nil, fmt.Errorf("error querying tokens by project_id '%s': %v", sourceID, err)
	}

	return mapRowsToDomainTokens(rows), nil
}
