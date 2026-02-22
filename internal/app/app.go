package app

import (
	"context"

	"github.com/tachyonhqdev/webhooks/internal/app/command"
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
	CreateEvent     CommandHandler[any]
	CreateEventType CommandHandler[command.CreateEventType]
	CreateExport    CommandHandler[any]
}

type Queries struct {
	EventTypes QueryHandler[any, any]
	Events     QueryHandler[any, any]
}

type App struct {
	Commands Commands
	Queries  Queries
}
