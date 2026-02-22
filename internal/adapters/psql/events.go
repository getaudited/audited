package psql

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/oklog/ulid/v2"
	"github.com/tachyonhqdev/webhooks/internal/adapters/models"
	"github.com/tachyonhqdev/webhooks/internal/domain"
)

type EventsPsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventsPsqlRepository(db boil.ContextExecutor) EventsPsqlRepository {
	return EventsPsqlRepository{db: db}
}

func (a EventsPsqlRepository) Save(ctx context.Context, e domain.Event) error {
	row := models.Event{
		ID:               e.Id,
		TenantID:         e.TenantID,
		Version:          e.Version,
		ActorID:          e.Actor.Id,
		ActorType:        e.Actor.ActorType,
		ActorName:        null.StringFromPtr(e.Actor.Name),
		ActorMetadata:    null.JSONFrom(e.Actor.Metadata),
		ContextLocation:  e.Context.Location,
		ContextUserAgent: null.StringFromPtr(e.Context.UserAgent),
		Metadata:         null.JSONFrom(e.Metadata),
		OccurredAt:       e.OccurredAt,
	}

	err := row.Insert(ctx, a.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("error saving event: %v", err)
	}

	targetRows := make([]*models.EventTarget, len(e.Targets))
	for i, target := range e.Targets {
		targetRows[i] = &models.EventTarget{
			InternalID: ulid.Make().String(),
			ID:         target.Id,
			EventID:    e.Id,
			Name:       null.StringFromPtr(target.Name),
			Type:       target.TargetType,
			Metadata:   null.JSONFrom(target.Metadata),
		}
	}

	err = row.AddEventTargets(ctx, a.db, true, targetRows...)
	if err != nil {
		return fmt.Errorf("error saving event_targets: %v", err)
	}

	return nil
}
