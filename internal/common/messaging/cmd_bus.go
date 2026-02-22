package messaging

import (
	"database/sql"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

/*func newCommandBus(publisher message.Publisher) (*cqrs.CommandBus, error) {
	return cqrs.NewCommandBusWithConfig(nil, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return "command." + params.CommandName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: nil,
	})
}*/

func NewTransactionalOutboxCommandBus(tx *sql.Tx, logger watermill.LoggerAdapter) (*cqrs.CommandBus, error) {
	publisher, err := NewCommandTransactionalOutboxPublisher(tx, logger)
	if err != nil {
		return nil, err
	}

	return cqrs.NewCommandBusWithConfig(publisher, cqrs.CommandBusConfig{
		GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger: logger,
	})
}
