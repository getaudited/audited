package psql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/friendsofgo/errors"
	"github.com/tachyonhqdev/webhooks/internal/adapters/models"
	"github.com/tachyonhqdev/webhooks/internal/domain"
)

type EventTypePsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventTypePsqlRepository(db boil.ContextExecutor) EventTypePsqlRepository {
	return EventTypePsqlRepository{db: db}
}

func (a EventTypePsqlRepository) Save(ctx context.Context, et domain.EventType) error {
	row := models.EventType{
		ID:                           et.Id,
		TenantID:                     et.TenantID,
		Version:                      et.Version,
		Action:                       et.Action,
		TargetTypes:                  et.TargetTypes,
		ShouldValidateMetadataSchema: et.ShouldValidateMetadataSchema,
		EventSchema:                  null.JSONFrom(et.Schema),
		CreatedAt:                    et.CreatedAt,
		UpdatedAt:                    et.UpdatedAt,
	}

	err := row.Insert(ctx, a.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("unable to save event type: %v", err)
	}

	return nil
}

func (a EventTypePsqlRepository) FindByAction(ctx context.Context, tenantID string, action string) (*domain.EventType, error) {
	row, err := models.EventTypes(models.EventTypeWhere.TenantID.EQ(tenantID), models.EventTypeWhere.Action.EQ(action)).One(ctx, a.db)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrEventTypeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error querying for event_type by action '%s' from tenant_id '%s': %v", action, tenantID, err)
	}

	return &domain.EventType{
		Id:                           row.ID,
		TenantID:                     row.TenantID,
		Version:                      row.Version,
		Action:                       row.Action,
		TargetTypes:                  row.TargetTypes, //,,
		ShouldValidateMetadataSchema: row.ShouldValidateMetadataSchema,
		Schema:                       nil, // TODO: fetch from event_type_schemas
		CreatedAt:                    row.CreatedAt,
		UpdatedAt:                    row.UpdatedAt,
	}, nil
}
