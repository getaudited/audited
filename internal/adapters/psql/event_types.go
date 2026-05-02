package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"

	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/domain"
)

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
	if err != nil {
		return fmt.Errorf("unable to save event type: %v", err)
	}

	return nil
}

func (r EventTypePsqlRepository) FindByAction(ctx context.Context, action string) (*domain.EventType, error) {
	row, err := models.EventTypes(models.EventTypeWhere.Action.EQ(action)).One(ctx, r.db)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrEventTypeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying for event_type by action '%s': %v", action, err)
	}

	return &domain.EventType{
		Id:                           row.ID,
		Version:                      row.Version,
		Action:                       row.Action,
		TargetTypes:                  row.TargetTypes, //,,
		ShouldValidateMetadataSchema: row.ShouldValidateMetadataSchema,
		Schema:                       nil, // TODO: fetch from event_type_schemas
		CreatedAt:                    row.CreatedAt,
		UpdatedAt:                    row.UpdatedAt,
	}, nil
}

func (r EventTypePsqlRepository) Delete(ctx context.Context, action string) error {
	_, err := models.EventTypes(models.EventTypeWhere.Action.EQ(action)).DeleteAll(ctx, r.db)
	if err != nil {
		return fmt.Errorf("error deleting event_type with action '%s' due to: %v", action, err)
	}

	return nil
}
