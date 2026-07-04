package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/lib/pq"

	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/domain"
)

const (
	constraintEventTypeActionIsUnique         = "pk_action"
	constraintOneEventTypeVersionPerEventType = "pk_event_type_id_and_version"
	minEventTypeVersion                       = 1
)

type EventTypePsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventTypePsqlRepository(db boil.ContextExecutor) EventTypePsqlRepository {
	return EventTypePsqlRepository{db: db}
}

func (r EventTypePsqlRepository) Save(ctx context.Context, et domain.EventType) error {
	row := models.EventType{
		Action:                       et.Action,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		CreatedAt:                    et.CreatedAt,
	}

	err := row.Insert(ctx, r.db, boil.Infer())
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Constraint == constraintEventTypeActionIsUnique {
		return domain.ErrEventTypeExists
	}
	if err != nil {
		return fmt.Errorf("unable to save event type: %w", err)
	}

	vr := mapEventTypeVersionToRow(et.Action, et.LastVersion)
	err = vr.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("error saving event_type_versions: %w", err)
	}

	return nil
}

func (r EventTypePsqlRepository) SaveVersion(
	ctx context.Context,
	action string,
	targetTypes []string,
	schema domain.Schema,
) error {
	lastEventTypeVersion, err := models.EventTypeVersions(
		models.EventTypeVersionWhere.EventTypeAction.EQ(action),
		qm.OrderBy("version DESC"),
		qm.For("UPDATE"),
	).One(ctx, r.db)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrEventTypeNotFound
	}
	if err != nil {
		return fmt.Errorf("unable to find the last event_type_version for '%s': %w", action, err)
	}

	row := &models.EventTypeVersion{
		EventTypeAction: action,
		Version:         lastEventTypeVersion.Version + 1,
		TargetTypes:     targetTypes,
		EventSchema:     null.JSONFrom(schema),
		CreatedAt:       time.Now(),
	}

	err = row.Insert(ctx, r.db, boil.Infer())
	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Constraint == constraintOneEventTypeVersionPerEventType {
		return domain.ErrEventTypeVersionExists
	}
	if err != nil {
		return fmt.Errorf("unable to save event_type_version: %w", err)
	}

	return nil
}

func (r EventTypePsqlRepository) RollbackVersion(ctx context.Context, action string) error {
	count, err := models.EventTypeVersions(models.EventTypeVersionWhere.EventTypeAction.EQ(action)).Count(ctx, r.db)
	if err != nil {
		return fmt.Errorf("unable to count event_type_versions by action '%s': %w", action, err)
	}

	if count == minEventTypeVersion {
		return domain.ErrVersionOneOfEventTypeCannotBeRolledBack
	}

	row, err := models.EventTypeVersions(
		models.EventTypeVersionWhere.EventTypeAction.EQ(action),
		qm.OrderBy("version DESC"),
	).One(ctx, r.db)
	if err != nil {
		return fmt.Errorf("unable to find the last event_type_version for '%s': %w", action, err)
	}

	_, err = row.Delete(ctx, r.db)
	if err != nil {
		return fmt.Errorf("unable to rollback event_type '%s': %w", action, err)
	}

	return nil
}

func (r EventTypePsqlRepository) FindByAction(ctx context.Context, action string) (query.EventType, error) {
	row, err := models.EventTypes(
		models.EventTypeWhere.Action.EQ(action),
		qm.Load(models.EventTypeRels.EventTypeActionEventTypeVersions),
	).One(ctx, r.db)
	if errors.Is(err, sql.ErrNoRows) {
		return query.EventType{}, domain.ErrEventTypeNotFound
	}
	if err != nil {
		return query.EventType{}, fmt.Errorf("error querying for event_type by action '%s': %w", action, err)
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

func (r EventTypePsqlRepository) QueryAll(ctx context.Context, params query.AllEventTypes) (query.Pagination[query.EventType], error) {
	count, err := models.EventTypes().Count(ctx, r.db)
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("unable to count event types: %w", err)
	}

	qms := []qm.QueryMod{
		qm.Limit(params.PaginationParams.Limit),
		qm.Offset(mapPaginationParamsToOffset(params.PaginationParams)),
		qm.OrderBy("created_at DESC"),
		qm.Load(models.EventTypeRels.EventTypeActionEventTypeVersions),
	}

	if params.Action != nil {
		qms = append(qms, models.EventTypeWhere.Action.ILIKE("%"+*params.Action+"%"))
	}

	rows, err := models.EventTypes(qms...).All(ctx, r.db)
	if err != nil {
		return query.Pagination[query.EventType]{}, fmt.Errorf("unable to query event types: %w", err)
	}

	return mapToPaginationResult[query.EventType](params.PaginationParams, count, mapRowsToEventTypes(rows)), nil
}

func (r EventTypePsqlRepository) AllVersionsByAction(ctx context.Context, action string) ([]query.EventType, error) {
	//TODO implement me
	panic("implement me")
}
