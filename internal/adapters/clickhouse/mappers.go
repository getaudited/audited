package clickhouse

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/getaudited/audited/internal/domain"
)

type EventRow struct {
	id                 string
	sourceID           string
	version            uint16
	action             string
	actorID            string
	actorType          string
	actorName          string
	actorMetadata      string
	contextLocation    string
	contextUserAgent   string
	metadata           string
	occurredAt         time.Time
	targetsInternalIDs []string
	targetsIds         []string
	targetsNames       []string
	targetsTargetTypes []string
	targetsMetadatas   []string
}

func mapRowToDomainEvent(row driver.Row) (domain.Event, error) {
	var er EventRow

	err := row.Scan(
		&er.id,
		&er.sourceID,
		&er.version,
		&er.action,
		&er.actorID,
		&er.actorType,
		&er.actorName,
		&er.actorMetadata,
		&er.contextLocation,
		&er.contextUserAgent,
		&er.metadata,
		&er.occurredAt,
		&er.targetsInternalIDs,
		&er.targetsIds,
		&er.targetsNames,
		&er.targetsTargetTypes,
		&er.targetsMetadatas,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Event{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.Event{}, fmt.Errorf("error querying event: %w", err)
	}

	eventTargets := make([]domain.Target, len(er.targetsIds))
	for i := 0; i < len(er.targetsIds); i++ {
		metadata, err := mapStringToMetadata(er.targetsMetadatas[i])
		if err != nil {
			return domain.Event{}, err
		}

		eventTargets[i] = domain.Target{
			ID:         er.targetsIds[i],
			Name:       new(er.targetsNames[i]),
			TargetType: er.targetsTargetTypes[i],
			Metadata:   &metadata,
		}
	}

	actorMetadata, err := mapStringToMetadata(er.actorMetadata)
	if err != nil {
		return domain.Event{}, err
	}

	metadata, err := mapStringToMetadata(er.metadata)
	if err != nil {
		return domain.Event{}, err
	}

	return domain.MarshallToEvent(
		er.id,
		er.sourceID,
		er.action,
		int(er.version),
		domain.Actor{
			ID:        er.actorID,
			ActorType: er.actorType,
			Name:      new(er.actorName),
			Metadata:  &actorMetadata,
		},
		eventTargets,
		domain.Context{
			Location:  er.contextLocation,
			UserAgent: new(er.contextUserAgent),
		},
		&metadata,
		er.occurredAt,
	), nil
}

func mapRowsToEvents(rows driver.Rows) ([]domain.Event, error) {
	var events []domain.Event

	for rows.Next() {
		event, err := mapRowToDomainEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func mapStringToMetadata(value string) (domain.Metadata, error) {
	var metadata domain.Metadata
	if strings.TrimSpace(value) != "" {
		err := json.Unmarshal([]byte(value), &metadata)
		if err != nil {
			return domain.Metadata{}, fmt.Errorf("error unmarshalling event target metadata: %w", err)
		}
	}

	return metadata, nil
}
