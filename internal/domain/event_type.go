package domain

import "time"

type EventType struct {
	Action                       string
	ShouldValidateMetadataSchema bool
	Version                      int
	TargetTypes                  []string
	Schema                       Schema
	LastVersion                  EventTypeVersion
	CreatedAt                    time.Time
}

type EventTypeVersion struct {
	Version     int
	TargetTypes []string
	Schema      Schema
	CreatedAt   time.Time
}

func NewEventTypeVersion(targetTypes []string, schema Schema) EventTypeVersion {
	return EventTypeVersion{
		Version:     1,
		TargetTypes: targetTypes,
		Schema:      schema,
		CreatedAt:   time.Now(),
	}
}
