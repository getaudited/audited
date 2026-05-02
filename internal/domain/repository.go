package domain

import "context"

type EventTypeRepository interface {
	Delete(ctx context.Context, action string) error
	Save(ctx context.Context, eventType EventType) error
}

type EventRepository interface {
	Save(ctx context.Context, evt Event, token TokenValue) error
}
