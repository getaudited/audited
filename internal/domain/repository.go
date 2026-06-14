package domain

import "context"

type EventTypeRepository interface {
	Delete(ctx context.Context, action string) error
	Save(ctx context.Context, eventType EventType) error
	RollbackVersion(ctx context.Context, action string) error
	SaveVersion(ctx context.Context, action string, targetTypes []string, schema Schema) error
}

type EventRepository interface {
	FindByID(ctx context.Context, id ID) (Event, error)
	Save(ctx context.Context, evt Event, token TokenValue) error
}
