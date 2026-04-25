package psql

import (
	"context"
	"database/sql"
	"fmt"
	"math"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/app/query"
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

func (s SourcesPsqlRepository) QueryAll(ctx context.Context, params query.PaginationParams) (query.Pagination[domain.Source], error) {
	count, err := models.Sources().Count(ctx, s.db)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error querying total sources: %w", err)
	}

	if count == 0 {
		return mapToPaginationResult[domain.Source](params, count, []domain.Source{}), nil
	}

	rows, err := models.Sources(
		qm.Limit(params.Limit),
		qm.Offset(mapPaginationParamsToOffset(params)),
		qm.OrderBy("created_at DESC"),
	).All(ctx, s.db)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error querying sources: %w", err)
	}

	return mapToPaginationResult[domain.Source](params, count, mapRowsToSources(rows)), nil
}

func mapPaginationParamsToOffset(params query.PaginationParams) int {
	if params.Page == 1 {
		return 0
	}

	return (params.Page - 1) * params.Limit
}

func mapToPaginationResult[T any](params query.PaginationParams, totalRows int64, data []T) query.Pagination[T] {
	return query.Pagination[T]{
		Data:        data,
		Total:       int(totalRows),
		PerPage:     params.Limit,
		CurrentPage: params.Page,
		TotalPages:  int(math.Ceil(float64(totalRows) / float64(params.Limit))),
	}
}

func mapRowsToSources(rows []*models.Source) []domain.Source {
	result := make([]domain.Source, len(rows))

	for i, row := range rows {
		result[i] = domain.MarshallToSource(row.ID, row.Name, row.CreatedAt, row.UpdatedAt)
	}

	return result
}
