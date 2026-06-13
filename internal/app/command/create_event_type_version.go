package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type CreateEventTypeVersion struct {
	Action      string
	TargetTypes []string
	Schema      domain.Schema
}

type CreateEventTypeVersionHandler struct {
	txProvider TransactionProvider
}

func NewCreateEventTypeVersionHandler(txProvider TransactionProvider) CreateEventTypeVersionHandler {
	return CreateEventTypeVersionHandler{
		txProvider: txProvider,
	}
}

func (c CreateEventTypeVersionHandler) Execute(ctx context.Context, cmd CreateEventTypeVersion) error {
	return c.txProvider.Transact(ctx, func(adapter TransactionAdapters) error {
		return adapter.EventTypeRepository.SaveVersion(ctx, cmd.Action, cmd.TargetTypes, cmd.Schema)
	})
}
