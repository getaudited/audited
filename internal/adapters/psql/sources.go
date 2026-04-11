package psql

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/firminochangani/audited/internal/domain"
)

type SourcesPsqlRepository struct {
	db *sql.DB
}

func NewSourcesPsqlRepository(db *sql.DB) *SourcesPsqlRepository {
	return &SourcesPsqlRepository{db: db}
}

func (s SourcesPsqlRepository) Save(ctx context.Context, source *domain.Source) error {
	row := mapDomainSourceToModelSource(source)
	err := row.Insert(ctx, s.db, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}
