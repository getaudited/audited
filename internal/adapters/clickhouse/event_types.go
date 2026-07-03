package clickhouse

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
	sq "github.com/Masterminds/squirrel"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
)

type EventTypesClickhouseRepository struct {
	db clickhousedb.Conn
}

func NewEventTypesClickhouseRepository(db clickhousedb.Conn) *EventTypesClickhouseRepository {
	return &EventTypesClickhouseRepository{
		db: db,
	}
}

func (r EventTypesClickhouseRepository) FindByAction(ctx context.Context, action string) (query.EventType, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT
					action,
					should_validate_metadata_schema,
					versions.version,
					versions.schema,
					versions.target_types,
					versions.created_at,
					created_at
				FROM event_types WHERE action = ?`,
		action,
	)

	return mapRowToEventType(row)
}

func (r EventTypesClickhouseRepository) QueryAll(ctx context.Context, params query.AllEventTypes) (query.Pagination[query.EventType], error) {
	var total uint64
	row := r.db.QueryRow(ctx, `SELECT COUNT(action) FROM event_types`)
	err := row.Scan(&total)
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("error counting event_types: %w", err)
	}

	queryAll := sq.
		Select(`
			action,
			should_validate_metadata_schema,
			versions.version,
			versions.schema,
			versions.target_types,
			versions.created_at,
			created_at
		`).
		From("event_types").
		Limit(uint64(params.PaginationParams.Limit)).
		Offset(uint64(mapPaginationParamsToOffset(params.PaginationParams)))

	if params.Action != nil {
		queryAll = queryAll.Where("ilike(action, ?)", "%"+*params.Action+"%")
	}

	q, args, err := queryAll.ToSql()
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("error building query: %w", err)
	}
	
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("error querying event_types: %w", err)
	}
	defer func() { _ = rows.Close() }()

	eventTypes, err := mapRowsToEventTypes(rows)
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("error mapping event_types: %w", err)
	}

	return mapToPaginationResult[query.EventType](params.PaginationParams, total, eventTypes), nil
}

func (r EventTypesClickhouseRepository) Delete(ctx context.Context, action string) error {
	err := r.db.Exec(ctx, `DELETE FROM event_types WHERE action = ?`, action)
	if err != nil {
		return fmt.Errorf("error deleting event_type '%s' due to: %w", action, err)
	}

	return nil
}

func (r EventTypesClickhouseRepository) Save(ctx context.Context, et domain.EventType) error {
	found, err := r.FindByAction(ctx, et.Action)
	if err != nil && !errors.Is(err, domain.ErrEventTypeNotFound) {
		return err
	}
	if found.Action == et.Action {
		return domain.ErrEventTypeExists
	}

	err = r.db.Exec(
		ctx,
		`
		INSERT INTO event_types (
			action,
			should_validate_metadata_schema,
			versions.version,
			versions.schema,
			versions.target_types,
			versions.created_at,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		et.Action,
		et.ShouldValidateMetadataSchema,
		[]uint16{uint16(et.LastVersion.Version)},
		[]string{string(et.LastVersion.Schema)},
		[][]string{et.LastVersion.TargetTypes},
		[]time.Time{et.LastVersion.CreatedAt},
		et.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error saving event_type: %w", err)
	}

	return nil
}

func (r EventTypesClickhouseRepository) RollbackVersion(ctx context.Context, action string) error {
	eventType, err := r.FindByAction(ctx, action)
	if err != nil {
		return err
	}
	if len(eventType.Versions) == 1 {
		return domain.ErrVersionOneOfEventTypeCannotBeRolledBack
	}

	err = r.db.Exec(ctx, `
		ALTER TABLE event_types
		UPDATE 
			versions.version = arrayFilter((v) -> v != arrayMax(versions.version), versions.version),
			versions.schema = arrayFilter((s, v) -> v != arrayMax(versions.version), versions.schema, versions.version),
			versions.target_types = arrayFilter((t, v) -> v != arrayMax(versions.version), versions.target_types, versions.version),
			versions.created_at = arrayFilter((c, v) -> v != arrayMax(versions.version), versions.created_at, versions.version)
		WHERE action = ?;
	`, action)
	if err != nil {
		return fmt.Errorf("error rolling back event_type version by action '%s' due to: %w", action, err)
	}

	return nil
}

func (r EventTypesClickhouseRepository) SaveVersion(ctx context.Context, action string, targetTypes []string, schema domain.Schema) error {
	eventType, err := r.FindByAction(ctx, action)
	if err != nil {
		return err
	}

	slices.SortFunc[[]query.EventTypeVersion](eventType.Versions, func(a query.EventTypeVersion, b query.EventTypeVersion) int {
		return cmp.Compare(b.Version, a.Version)
	})

	lastEventTypeVersion := eventType.Versions[len(eventType.Versions)-1]

	err = r.db.Exec(
		ctx,
		`
		ALTER TABLE event_types
		UPDATE
			versions.version = arrayPushBack(versions.version, ?),
			versions.schema = arrayPushBack(versions.schema, ?),
			versions.target_types = arrayPushBack(versions.target_types, ?),
			versions.created_at = arrayPushBack(versions.created_at, ?)
		WHERE action = ?;
	`,
		lastEventTypeVersion.Version+1,
		string(schema),
		targetTypes,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error saving new version of event_type '%s' due to: %w", action, err)
	}

	return nil
}
