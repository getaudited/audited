package messaging

import (
	"database/sql"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
)

const commandsForwardedSqlTopic = "whk_commands_forwarded"

func newTransactionalOutboxPublisher(tx *sql.Tx, topic string, logger watermill.LoggerAdapter) (message.Publisher, error) {
	var err error
	var publisher message.Publisher
	publisher, err = watermillsql.NewPublisher(
		tx,
		watermillsql.PublisherConfig{
			SchemaAdapter: watermillsql.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating publisher: %v", err)
	}

	publisher = forwarder.NewPublisher(publisher, forwarder.PublisherConfig{
		ForwarderTopic: topic,
	})

	return publisher, nil
}

// NewCommandTransactionalOutboxPublisher creates an SQL publisher decorated with a forwarder publisher to store every command
// and then forward them to be
func NewCommandTransactionalOutboxPublisher(tx *sql.Tx, logger watermill.LoggerAdapter) (message.Publisher, error) {
	return newTransactionalOutboxPublisher(tx, commandsForwardedSqlTopic, logger)
}
