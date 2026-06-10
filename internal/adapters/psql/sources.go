package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
	"github.com/lib/pq"
)

const ConstraintSourceNameIsUnique = "un_source_name"

type SourcesPsqlRepository struct {
	db *sql.DB
}

func NewSourcesPsqlRepository(db *sql.DB) *SourcesPsqlRepository {
	return &SourcesPsqlRepository{db: db}
}

func (s SourcesPsqlRepository) Save(ctx context.Context, source *domain.Source) error {
	row := mapDomainSourceToModelSource(source)
	err := row.Insert(ctx, s.db, boil.Infer())
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Constraint == ConstraintSourceNameIsUnique {
		return domain.ErrSourceWithProvidedNameExists
	}
	if err != nil {
		return err
	}

	return nil
}

func (s SourcesPsqlRepository) FindByID(ctx context.Context, id string) (*domain.Source, error) {
	row, err := models.Sources(models.SourceWhere.ID.EQ(id)).One(ctx, s.db)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrSourceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying source by id '%s': %w", id, err)
	}

	return new(domain.MarshallToSource(row.ID, row.Name, row.CreatedAt, row.UpdatedAt)), nil
}

func (s SourcesPsqlRepository) QueryAll(ctx context.Context, params query.AllSources) (query.Pagination[domain.Source], error) {
	count, err := models.Sources().Count(ctx, s.db)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error querying total sources: %w", err)
	}

	if count == 0 {
		return mapToPaginationResult[domain.Source](params.Pagination, count, []domain.Source{}), nil
	}

	qms := []qm.QueryMod{
		qm.Limit(params.Pagination.Limit),
		qm.Offset(mapPaginationParamsToOffset(params.Pagination)),
		qm.OrderBy("created_at DESC"),
	}

	if params.Name != nil {
		qms = append(qms, models.SourceWhere.Name.ILIKE("%"+*params.Name+"%"))
	}

	rows, err := models.Sources(qms...).All(ctx, s.db)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error querying sources: %w", err)
	}

	return mapToPaginationResult[domain.Source](params.Pagination, count, mapRowsToSources(rows)), nil
}
