package psql

import (
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/firminochangani/audited/internal/adapters/models"
	"github.com/firminochangani/audited/internal/domain"
	"github.com/oklog/ulid/v2"
)

func mapDomainEventTargetsToModelEventTargets(eventID domain.ID, targets []domain.Target) ([]*models.EventTarget, error) {
	targetRows := make([]*models.EventTarget, len(targets))

	for i, target := range targets {
		targetMetadata, err := mapMetadataToJSON(target.Metadata)
		if err != nil {
			return nil, err
		}

		targetRows[i] = &models.EventTarget{
			InternalID: ulid.Make().String(),
			ID:         target.ID,
			EventID:    eventID.String(),
			Name:       null.StringFromPtr(target.Name),
			Type:       target.TargetType,
			Metadata:   targetMetadata,
		}
	}

	return targetRows, nil
}

func mapDomainEventToModelEvent(e domain.Event) (models.Event, error) {
	actorMetadata, err := mapMetadataToJSON(e.Actor().Metadata)
	if err != nil {
		return models.Event{}, err
	}

	eventMetadata, err := mapMetadataToJSON(e.Metadata())
	if err != nil {
		return models.Event{}, err
	}

	row := models.Event{
		ID:               e.ID().String(),
		SourceID:         e.SourceID().String(),
		Version:          e.Version(),
		ActorID:          e.Actor().ID,
		ActorType:        e.Actor().ActorType,
		ActorName:        null.StringFromPtr(e.Actor().Name),
		ActorMetadata:    actorMetadata,
		ContextLocation:  e.Context().Location,
		ContextUserAgent: null.StringFromPtr(e.Context().UserAgent),
		Metadata:         eventMetadata,
		OccurredAt:       e.OccurredAt(),
	}

	return row, err
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

func mapDomainSourceToModelSource(source *domain.Source) *models.Source {
	return &models.Source{
		ID:        source.ID().String(),
		Name:      source.Name(),
		CreatedAt: source.CreatedAt(),
		UpdatedAt: source.UpdatedAt(),
	}
}

func mapDomainTokenToModelToken(token domain.Token) *models.Token {
	return &models.Token{
		ID:        token.ID().String(),
		SourceID:  token.SourceID().String(),
		Name:      token.Name(),
		Value:     token.Value().String(),
		CreatedAt: token.CreatedAt(),
	}
}
