package command

import (
	"context"

	"github.com/getaudited/audited/internal/domain"
)

type TransactionAdapters struct {
	EventTypeRepository domain.EventTypeRepository
}

type TransactFunc func(adapter TransactionAdapters) error

type TransactionProvider interface {
	Transact(ctx context.Context, f TransactFunc) error
}
