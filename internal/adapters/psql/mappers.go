package psql

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/getaudited/audited/internal/adapters/models"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
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
		Action:           e.Action(),
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
		return null.JSON{}, fmt.Errorf("error mapping metadata into json: %w", err)
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

func mapDomainTokenToModelToken(token *domain.Token) *models.Token {
	return &models.Token{
		ID:        token.ID().String(),
		Name:      token.Name(),
		Value:     token.Value().String(),
		SourceID:  token.SourceID().String(),
		CreatedAt: token.CreatedAt(),
	}
}

func unmarshallCursor(encodedCursor string) (Cursor, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return Cursor{}, fmt.Errorf("error decoding cursor '%s': %w", encodedCursor, err)
	}

	var cursor Cursor
	err = json.Unmarshal(decodedCursor, &cursor)
	if err != nil {
		return Cursor{}, fmt.Errorf("error unmarshalling cursor '%s': %w", encodedCursor, err)
	}

	return cursor, nil
}

func marshallCursor(eventID string, occurredAt time.Time) (string, error) {
	cursor := Cursor{
		EventID:    eventID,
		OccurredAt: occurredAt,
	}

	marshalledCursor, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("error marshalling cursor: %w", err)
	}

	return base64.StdEncoding.EncodeToString(marshalledCursor), nil
}

func mapJsonToDomainMetadata(jsonMetadata null.JSON) (domain.Metadata, error) {
	var metadata domain.Metadata

	if jsonMetadata.IsZero() {
		return metadata, nil
	}

	err := json.Unmarshal(jsonMetadata.JSON, &metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %w", err)
	}

	return metadata, nil
}

func mapRowsToDomainEvents(rows []*models.Event) ([]domain.Event, error) {
	events := make([]domain.Event, len(rows))

	for i, row := range rows {
		metadata, err := mapJsonToDomainMetadata(row.Metadata)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling event metadata: %w", err)
		}

		actorMetadata, err := mapJsonToDomainMetadata(row.ActorMetadata)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling actor metadata: %w", err)
		}

		var targets []domain.Target
		if row.R != nil {
			targets, err = mapRowsToDomainTargets(row.R.EventTargets)
			if err != nil {
				return nil, err
			}
		}

		events[i] = domain.MarshallToEvent(
			row.ID,
			row.SourceID,
			row.Action,
			row.Version,
			domain.Actor{
				ID:        row.ActorID,
				ActorType: row.ActorType,
				Name:      row.ActorName.Ptr(),
				Metadata:  &actorMetadata,
			},
			targets,
			domain.Context{
				Location:  row.ContextLocation,
				UserAgent: row.ContextUserAgent.Ptr(),
			},
			&metadata,
			row.OccurredAt,
		)
	}

	return events, nil
}

func mapRowsToDomainTargets(rows []*models.EventTarget) ([]domain.Target, error) {
	targets := make([]domain.Target, len(rows))

	for i, row := range rows {
		metadata, err := mapJsonToDomainMetadata(row.Metadata)
		if err != nil {
			return nil, err
		}

		targets[i] = domain.Target{
			ID:         row.ID,
			Name:       row.Name.Ptr(),
			TargetType: row.Type,
			Metadata:   &metadata,
		}
	}

	return targets, nil
}

func mapToLimit(limit *int) int {
	if limit == nil {
		return 50
	}

	return *limit
}

func mapLastItemCursor(rows []*models.Event) (string, error) {
	if len(rows) == 0 {
		return "", nil
	}

	lastRow := rows[len(rows)-1]
	return marshallCursor(lastRow.ID, lastRow.OccurredAt)
}

func mapRowsToDomainTokens(rows []*models.Token) []*domain.Token {
	r := make([]*domain.Token, len(rows))

	for i, row := range rows {
		r[i] = domain.MarshallToToken(row.ID, row.SourceID, row.Value, row.Name, row.CreatedAt)
	}

	return r
}

func mapPaginationParamsToOffset(params query.PaginationParams) int {
	if params.Page == 1 {
		return 0
	}

	return (params.Page - 1) * params.Limit
}

func mapToPaginationResult[T any](params query.PaginationParams, totalRows int64, data []T) query.Pagination[T] {
	return query.Pagination[T]{
		Data:        data,
		Total:       int(totalRows),
		PerPage:     params.Limit,
		CurrentPage: params.Page,
		TotalPages:  int(math.Ceil(float64(totalRows) / float64(params.Limit))),
	}
}

func mapRowsToSources(rows []*models.Source) []domain.Source {
	result := make([]domain.Source, len(rows))

	for i, row := range rows {
		result[i] = domain.MarshallToSource(row.ID, row.Name, row.CreatedAt, row.UpdatedAt)
	}

	return result
}

func mapRowToEventType(row *models.EventType) *domain.EventType {
	return &domain.EventType{
		Id:                           row.ID,
		Version:                      row.Version,
		Action:                       row.Action,
		TargetTypes:                  row.TargetTypes,
		ShouldValidateMetadataSchema: row.ShouldValidateMetadataSchema,
		Schema:                       row.EventSchema.JSON,
		CreatedAt:                    row.CreatedAt,
		UpdatedAt:                    row.UpdatedAt,
	}
}

func mapRowsToEventTypes(rows []*models.EventType) []*domain.EventType {
	result := make([]*domain.EventType, len(rows))

	for i, row := range rows {
		result[i] = mapRowToEventType(row)
	}

	return result
}
