package domain

import (
	"errors"
	"strings"
	"time"
)

type EventType struct {
	action                       string
	shouldValidateMetadataSchema bool
	version                      int
	targetTypes                  []string
	schema                       Schema
	createdAt                    time.Time
}

func NewEventType(action string, validateSchema bool, targetTypes []string, schema Schema) (*EventType, error) {
	if strings.TrimSpace(action) == "" {
		return nil, errors.New("action cannot be empty")
	}

	return &EventType{
		action:                       action,
		shouldValidateMetadataSchema: validateSchema,
		version:                      1,
		targetTypes:                  deduplicateTargetTypes(targetTypes),
		schema:                       schema,
		createdAt:                    time.Now(),
	}, nil
}

func (e *EventType) Action() string {
	return e.action
}

func (e *EventType) ShouldValidateMetadataSchema() bool {
	return e.shouldValidateMetadataSchema
}

func (e *EventType) Version() int {
	return e.version
}

func (e *EventType) TargetTypes() []string {
	return e.targetTypes
}

func (e *EventType) Schema() Schema {
	return e.schema
}

func (e *EventType) CreatedAt() time.Time {
	return e.createdAt
}

/*func MarshallToEventType(
	action string,
	version int,
	validateSchema bool,
	targetTypes []string,
	schema Schema,
	createdAt time.Time,
) *EventType {
	return &EventType{
		action:                       action,
		shouldValidateMetadataSchema: validateSchema,
		version:                      version,
		targetTypes:                  targetTypes,
		schema:                       schema,
		createdAt:                    createdAt,
	}
}*/

func deduplicateTargetTypes(targetTypes []string) []string {
	var result []string
	mappedTargetTypes := map[string]struct{}{}

	for _, tt := range targetTypes {
		_, exists := mappedTargetTypes[tt]
		if exists {
			continue
		}

		mappedTargetTypes[tt] = struct{}{}
		result = append(result, tt)
	}

	return result
}
