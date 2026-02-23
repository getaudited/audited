package app

import (
	"context"

	"github.com/tachyonhqdev/webhooks/internal/app/command"
	"github.com/tachyonhqdev/webhooks/internal/app/query"
	"github.com/tachyonhqdev/webhooks/internal/domain"
)

type CommandHandler[C any] interface {
	Execute(ctx context.Context, cmd C) error
}

type CommandWithResultHandler[C, R any] interface {
	Execute(ctx context.Context, cmd C) (R, error)
}

type QueryHandler[Q, R any] interface {
	Execute(ctx context.Context, q Q) (R, error)
}

type Commands struct {
	CreateEvent     CommandHandler[command.CreateEvent]
	CreateEventType CommandHandler[command.CreateEventType]
	CreateExport    CommandHandler[any]
}

type Queries struct {
	EventTypes        QueryHandler[any, query.Pagination[domain.EventType]]
	EventTypeByAction QueryHandler[query.EventTypeByAction, *domain.EventType]
	Events            QueryHandler[any, query.Pagination[domain.Event]]
	EventByID         QueryHandler[any, domain.Event]
}

type App struct {
	Commands Commands
	Queries  Queries
}
