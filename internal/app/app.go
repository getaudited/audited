package app

import (
	"context"

	"github.com/firminochangani/audited/internal/app/command"
	"github.com/firminochangani/audited/internal/app/query"
	"github.com/firminochangani/audited/internal/domain"
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
	DeleteEventType CommandHandler[command.DeleteEventType]
	CreateExport    CommandHandler[any]
	CreateSource    CommandHandler[command.CreateSource]
	CreateToken     CommandHandler[command.CreateToken]
	DeleteToken     CommandHandler[command.DeleteToken]
}

type Queries struct {
	EventTypes        QueryHandler[any, query.Pagination[domain.EventType]]
	EventTypeByAction QueryHandler[query.EventTypeByName, *domain.EventType]
	Events            QueryHandler[query.AllEvents, query.CursorPaginationResult[domain.Event]]
	EventByID         QueryHandler[any, domain.Event]
	SourceByID        QueryHandler[query.SourceByID, *domain.Source]
	AllSources        QueryHandler[query.AllSources, query.Pagination[domain.Source]]
	AllTokens         QueryHandler[query.AllTokens, []*domain.Token]
	AllEventTypes     QueryHandler[query.AllEventTypes, query.Pagination[*domain.EventType]]
}

type App struct {
	Commands Commands
	Queries  Queries
}
