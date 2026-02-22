package domain

import "time"

type EventType struct {
	Id                           string
	TenantID                     string
	Version                      int
	Action                       string
	TargetTypes                  []string
	ShouldValidateMetadataSchema bool
	Schema                       Schema
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}
