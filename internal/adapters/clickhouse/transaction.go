package clickhouse

import (
	"context"

	clickhousedb "github.com/ClickHouse/clickhouse-go/v2"

	"github.com/getaudited/audited/internal/app/command"
)

// ShallowTxProvider it implements the interface TransactionProvider, but it doesn't actually begin a transaction since
// Clickhouse does not support it natively.
type ShallowTxProvider struct {
	db clickhousedb.Conn
}

func NewShallowTxProvider(db clickhousedb.Conn) *ShallowTxProvider {
	return &ShallowTxProvider{
		db: db,
	}
}

func (t ShallowTxProvider) Transact(_ context.Context, f command.TransactFunc) error {
	adapters := command.TransactionAdapters{
		EventTypeRepository: NewEventTypesClickhouseRepository(t.db),
	}

	return f(adapters)
}
