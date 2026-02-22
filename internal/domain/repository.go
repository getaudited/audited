package domain

import "context"

type EventTypeRepository interface {
	Save(ctx context.Context, evt EventType) error
}
