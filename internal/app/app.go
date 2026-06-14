package app

import (
	"context"

	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/app/query"
	"github.com/getaudited/audited/internal/domain"
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
	CreateEvent              CommandHandler[command.CreateEvent]
	CreateEventType          CommandHandler[command.CreateEventType]
	CreateEventTypeVersion   CommandHandler[command.CreateEventTypeVersion]
	RollbackEventTypeVersion CommandHandler[command.RollbackEventTypeVersion]
	DeleteEventType          CommandHandler[command.DeleteEventType]
	CreateExport             CommandHandler[any]
	CreateSource             CommandHandler[command.CreateSource]
	CreateToken              CommandHandler[command.CreateToken]
	DeleteToken              CommandHandler[command.DeleteToken]
	LogIn                    CommandWithResultHandler[command.LogIn, string]
	CreateAdminUser          CommandHandler[command.CreateAdminUser]
}

type Queries struct {
	EventTypeByAction QueryHandler[query.EventTypeByAction, query.EventType]
	Events            QueryHandler[query.AllEvents, query.CursorPaginationResult[domain.Event]]
	EventByID         QueryHandler[query.EventByID, domain.Event]
	SourceByID        QueryHandler[query.SourceByID, *domain.Source]
	AllSources        QueryHandler[query.AllSources, query.Pagination[domain.Source]]
	AllTokens         QueryHandler[query.AllTokens, []*domain.Token]
	AllEventTypes     QueryHandler[query.AllEventTypes, query.Pagination[query.EventType]]
	UserProfile       QueryHandler[query.UserProfile, *domain.User]
}

type App struct {
	Commands Commands
	Queries  Queries
}
