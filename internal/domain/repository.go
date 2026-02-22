package domain

import "context"

type EventTypeRepository interface {
	Save(ctx context.Context, evt EventType) error
}

type EventRepository interface {
	Save(ctx context.Context, evt Event) error
}
