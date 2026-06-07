package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/lib/pq"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/domain"
)

const ConstraintEventTypeActionIsUnique = "un_event_type_name"

type EventTypePsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventTypePsqlRepository(db boil.ContextExecutor) EventTypePsqlRepository {
	return EventTypePsqlRepository{db: db}
}

func (r EventTypePsqlRepository) Save(ctx context.Context, et domain.EventType) error {
	row := models.EventType{
		ID:                           et.Id,
		Version:                      et.Version,
		Action:                       et.Action,
		TargetTypes:                  et.TargetTypes,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		EventSchema:                  null.JSONFrom(et.Schema),
		CreatedAt:                    et.CreatedAt,
		UpdatedAt:                    et.UpdatedAt,
	}

	err := row.Insert(ctx, r.db, boil.Infer())
	fmt.Println("err", err)
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Constraint == ConstraintEventTypeActionIsUnique {
		return domain.ErrEventTypeExists
	}
	if err != nil {
		return fmt.Errorf("unable to save event type: %w", err)
	}

	return nil
}

func (r EventTypePsqlRepository) FindByAction(ctx context.Context, action string) (*domain.EventType, error) {
	row, err := models.EventTypes(models.EventTypeWhere.Action.EQ(action)).One(ctx, r.db)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrEventTypeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying for event_type by action '%s': %w", action, err)
	}

	return mapRowToEventType(row), nil
}

func (r EventTypePsqlRepository) Delete(ctx context.Context, action string) error {
	_, err := models.EventTypes(models.EventTypeWhere.Action.EQ(action)).DeleteAll(ctx, r.db)
	if err != nil {
		return fmt.Errorf("error deleting event_type with action '%s' due to: %w", action, err)
	}

	return nil
}

func (r EventTypePsqlRepository) QueryAll(ctx context.Context, params query.PaginationParams) (query.Pagination[*domain.EventType], error) {
	count, err := models.EventTypes().Count(ctx, r.db)
	if err != nil {
		return query.Pagination[*domain.EventType]{}, fmt.Errorf("unable to count event types: %w", err)
	}
	boil.DebugMode = true
	rows, err := models.EventTypes(
		qm.Limit(params.Limit),
		qm.Offset(mapPaginationParamsToOffset(params)),
		qm.OrderBy("created_at DESC"),
	).All(ctx, r.db)
	if err != nil {
		return query.Pagination[*domain.EventType]{}, fmt.Errorf("unable to query event types: %w", err)
	}

	fmt.Println("###", rows)

	return mapToPaginationResult[*domain.EventType](params, count, mapRowsToEventTypes(rows)), nil
}
