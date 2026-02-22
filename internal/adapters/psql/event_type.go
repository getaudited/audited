package psql

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/tachyonhqdev/webhooks/internal/adapters/models"
	"github.com/tachyonhqdev/webhooks/internal/domain"
)

type EventTypePsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventTypePsqlRepository(db boil.ContextExecutor) EventTypePsqlRepository {
	return EventTypePsqlRepository{db: db}
}

func (a EventTypePsqlRepository) Save(ctx context.Context, evt domain.EventType) error {
	row := models.EventType{
		ID:          evt.Id,
		TenantID:    evt.TenantID,
		Action:      evt.Action,
		EventSchema: null.JSONFrom(evt.Schema),
		CreatedAt:   evt.CreatedAt,
		UpdatedAt:   evt.UpdatedAt,
	}

	err := row.Insert(ctx, a.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("unable to save event type: %v", err)
	}

	return nil
}
