package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	sq "github.com/Masterminds/squirrel"

	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

type SourcesClickhouseRepository struct {
	db clickhousedb.Conn
}

func NewSourcesClickhouseRepository(db clickhousedb.Conn) SourcesClickhouseRepository {
	return SourcesClickhouseRepository{
		db: db,
	}
}

func (r SourcesClickhouseRepository) Save(ctx context.Context, s *domain.Source) error {
	err := r.db.Exec(
		ctx,
		`INSERT INTO sources (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		s.ID().String(),
		s.Name(),
		s.CreatedAt(),
		s.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("error saving source '%s' due to: %w", s.Name(), err)
	}

	return nil
}

func (r SourcesClickhouseRepository) QueryAll(
	ctx context.Context,
	params query.AllSources,
) (query.Pagination[domain.Source], error) {
	var total uint64
	row := r.db.QueryRow(ctx, `SELECT COUNT(id) FROM sources`)
	err := row.Scan(&total)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error counting sources: %w", err)
	}

	queryAll := sq.
		Select("id, name, created_at, updated_at").
		From("sources").
		Limit(uint64(params.Pagination.Limit)).
		Offset(uint64(mapPaginationParamsToOffset(params.Pagination)))

	if params.Name != nil {
		queryAll = queryAll.Where("ilike(name, ?)", "%"+*params.Name+"%")
	}

	q, args, err := queryAll.ToSql()
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error building query: %w", err)
	}

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error querying sources: %w", err)
	}
	defer func() { _ = rows.Close() }()

	sources, err := mapRowsToSources(rows)
	if err != nil {
		return query.Pagination[domain.Source]{}, fmt.Errorf("error mapping sources: %w", err)
	}

	return mapToPaginationResult[domain.Source](params.Pagination, total, sources), nil
}

func (r SourcesClickhouseRepository) FindByID(ctx context.Context, id string) (*domain.Source, error) {
	row := r.db.QueryRow(ctx, `SELECT id, name, created_at, updated_at FROM sources WHERE id = ?`, id)
	source, err := mapRowToSource(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrSourceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying source by id '%s' due to: %w", id, err)
	}

	return source, nil
}
