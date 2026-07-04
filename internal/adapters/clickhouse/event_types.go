package clickhouse

import (
	"context"
	"fmt"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"
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
					version,
					schema,
					target_types,
					created_at
				FROM event_types
				WHERE action = ?
				ORDER BY version DESC
				LIMIT 1`,
		action,
	)

	return mapRowToEventType(row)
}

func (r EventTypesClickhouseRepository) QueryAll(ctx context.Context, params query.AllEventTypes) (query.Pagination[query.EventType], error) {
	var total uint64
	row := r.db.QueryRow(ctx, `SELECT COUNT(DISTINCT(action)) FROM event_types`)
	err := row.Scan(&total)
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("error counting event_types: %w", err)
	}

	q := `
		SELECT * FROM (
		    SELECT
				action,
				should_validate_metadata_schema,
				version,
				schema,
				target_types,
				created_at
			FROM event_types
			ORDER BY action ASC, version DESC
			LIMIT 1 BY action
		)
	`

	var args []any
	if params.Action != nil {
		q += " WHERE ilike(action, ?)"
		args = append(args, "%"+*params.Action+"%")
	}

	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, params.PaginationParams.Limit, mapPaginationParamsToOffset(params.PaginationParams))

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
	//row := r.db.QueryRow(ctx, `SELECT action FROM event_types WHERE action = ?`, et.Action)

	// found, err := r.FindByAction(ctx, et.Action)
	// if err != nil && !errors.Is(err, domain.ErrEventTypeNotFound) {
	// 	return err
	// }
	// if found.Action == et.Action {
	// 	return domain.ErrEventTypeExists
	// }

	err := r.db.Exec(
		ctx,
		`
		INSERT INTO event_types (
			action,
			should_validate_metadata_schema,
			version,
			schema,
			target_types,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		et.Action,
		et.ShouldValidateMetadataSchema,
		uint16(et.LastVersion.Version),
		string(et.LastVersion.Schema),
		et.LastVersion.TargetTypes,
		et.LastVersion.CreatedAt,
		et.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error saving event_type: %w", err)
	}

	return nil
}

func (r EventTypesClickhouseRepository) RollbackVersion(ctx context.Context, action string) error {
	evt, err := r.FindByAction(ctx, action)
	if err != nil {
		return err
	}
	if evt.Version == 1 {
		return domain.ErrVersionOneOfEventTypeCannotBeRolledBack
	}

	err = r.db.Exec(ctx, `DELETE FROM event_types WHERE action = ? AND version = ?`, action, evt.Version)
	if err != nil {
		return fmt.Errorf("error rolling back event_type version by action '%s' due to: %w", action, err)
	}

	return nil
}

func (r EventTypesClickhouseRepository) SaveVersion(ctx context.Context, action string, targetTypes []string, schema domain.Schema) error {
	et, err := r.FindByAction(ctx, action)
	if err != nil {
		return err
	}

	err = r.db.Exec(
		ctx,
		`
		INSERT INTO event_types (
			action,
			should_validate_metadata_schema,
			version,
			schema,
			target_types,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		action,
		et.ShouldValidateMetadataSchema,
		uint16(et.Version+1),
		string(schema),
		targetTypes,
		et.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error saving new version of event_type '%s' due to: %w", action, err)
	}

	return nil
}

func (r EventTypesClickhouseRepository) AllVersionsByAction(ctx context.Context, action string) ([]query.EventType, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			action,
			should_validate_metadata_schema,
			version,
			schema,
			target_types,
			created_at
		FROM event_types
		WHERE action = ?
		ORDER BY version DESC
	`, action)
	if err != nil {
		return nil, fmt.Errorf("error querying all versions of event '%s' due to: %w", action, err)
	}

	return mapRowsToEventTypes(rows)
}
