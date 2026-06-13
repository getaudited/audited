package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/getaudited/audited/internal/app/command"
	"github.com/getaudited/audited/internal/common/logs"
)

type TxProvider struct {
	db     *sql.DB
	logger *logs.Logger
}

func NewTxProvider(db *sql.DB, logger *logs.Logger) *TxProvider {
	return &TxProvider{
		db:     db,
		logger: logger,
	}
}

func (t TxProvider) Transact(ctx context.Context, f command.TransactFunc) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %w", err)
	}

	adapters := command.TransactionAdapters{
		EventTypeRepository: NewEventTypePsqlRepository(tx),
	}

	err = f(adapters)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			t.logger.Error("error while rolling back transaction", "error", err)
			return errors.Join(err, rollbackErr)
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to commit transaction: %w", err)
	}

	return nil
}
