package psql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/oklog/ulid/v2"
)

type EventsPsqlRepository struct {
	db boil.ContextExecutor
}

func NewEventsPsqlRepository(db boil.ContextExecutor) EventsPsqlRepository {
	return EventsPsqlRepository{db: db}
}

func (a EventsPsqlRepository) Save(ctx context.Context, e domain.Event) error {
	actorMetadata, err := mapMetadataToJSON(e.Actor.Metadata)
	if err != nil {
		return err
	}

	eventMetadata, err := mapMetadataToJSON(e.Metadata)
	if err != nil {
		return err
	}

	row := models.Event{
		ID:               e.Id,
		TenantID:         e.TenantID,
		Version:          e.Version,
		ActorID:          e.Actor.Id,
		ActorType:        e.Actor.ActorType,
		ActorName:        null.StringFromPtr(e.Actor.Name),
		ActorMetadata:    actorMetadata,
		ContextLocation:  e.Context.Location,
		ContextUserAgent: null.StringFromPtr(e.Context.UserAgent),
		Metadata:         eventMetadata,
		OccurredAt:       e.OccurredAt,
	}

	err = row.Insert(ctx, a.db, boil.Infer())
	if err != nil {
		return fmt.Errorf("error saving event: %v", err)
	}

	targetRows := make([]*models.EventTarget, len(e.Targets))
	for i, target := range e.Targets {
		targetMetadata, err := mapMetadataToJSON(target.Metadata)
		if err != nil {
			return err
		}

		targetRows[i] = &models.EventTarget{
			InternalID: ulid.Make().String(),
			ID:         target.Id,
			EventID:    e.Id,
			Name:       null.StringFromPtr(target.Name),
			Type:       target.TargetType,
			Metadata:   targetMetadata,
		}
	}

	err = row.AddEventTargets(ctx, a.db, true, targetRows...)
	if err != nil {
		return fmt.Errorf("error saving event_targets: %v", err)
	}

	return nil
}

func mapMetadataToJSON(metadata *domain.Metadata) (null.JSON, error) {
	if metadata == nil {
		return null.JSON{}, nil
	}

	data, err := json.Marshal(*metadata)
	if err != nil {
		return null.JSON{}, fmt.Errorf("error mapping metadata into json: %v", err)
	}

	return null.JSONFrom(data), nil
}
