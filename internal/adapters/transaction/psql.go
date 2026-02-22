package transaction

import (
	"context"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/tachyonhqdev/webhooks/internal/adapters/command_bus"

	"github.com/tachyonhqdev/webhooks/internal/adapters/psql"
	"github.com/tachyonhqdev/webhooks/internal/app/command"
	"github.com/tachyonhqdev/webhooks/internal/common/logs"
	messaginglib "github.com/tachyonhqdev/webhooks/internal/common/messaging"
)

type PsqlProvider struct {
	logger    *logs.Logger
	db        boil.ContextBeginner
	messaging *messaginglib.Messaging
}

func NewPsqlProvider(db boil.ContextBeginner, messaging *messaginglib.Messaging, logger *logs.Logger) PsqlProvider {
	return PsqlProvider{
		db:        db,
		logger:    logger,
		messaging: messaging,
	}
}

func (p PsqlProvider) Transact(ctx context.Context, f command.TransactFunc) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to begin transaction: %v", err)
	}

	txCmdBus, err := messaginglib.NewTransactionalOutboxCommandBus(tx, p.messaging.Logger())
	if err != nil {
		return err
	}

	adapters := command.TransactableAdapters{
		EndpointRepository: psql.NewEndpointPsqlRepository(tx),
		MessageRepository:  psql.NewMessagePsqlRepository(tx),
		CommandBus:         command_bus.NewCommandSender(txCmdBus),
	}

	if err = f(adapters); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			p.logger.Error("Rollback error", "rollback_err", rollbackErr)
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
