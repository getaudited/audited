package messaging

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	watermillsql "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/firminochangani/audited/internal/common/logs"
)

const (
	maxMessagesPerSecond = 100
	poisonQueueName      = "watermill_poison_queue"
)

type Messaging struct {
	amqpURL          string
	baseLogger       *logs.Logger
	logger           watermill.LoggerAdapter
	router           *message.Router
	commandProcessor *cqrs.CommandProcessor
	commandForwarder *forwarder.Forwarder
}

func NewMessaging(db *sql.DB, amqpURL string, baseLogger *logs.Logger) (*Messaging, error) {
	logger := watermill.NewSlogLogger(baseLogger.Logger)

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating router: %v", err)
	}

	m := &Messaging{
		logger:     logger,
		router:     router,
		amqpURL:    amqpURL,
		baseLogger: baseLogger,
	}

	publisher, err := amqp.NewPublisher(amqp.NewDurableQueueConfig(amqpURL), logger)
	if err != nil {
		return nil, err
	}

	err = m.registerMiddlewares(publisher)
	if err != nil {
		return nil, err
	}

	cmdForwarderSubscriber, err := watermillsql.NewSubscriber(
		db,
		watermillsql.SubscriberConfig{
			SchemaAdapter:  watermillsql.DefaultPostgreSQLSchema{},
			OffsetsAdapter: watermillsql.DefaultPostgreSQLOffsetsAdapter{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating subscriber for forwarder: %v", err)
	}

	err = cmdForwarderSubscriber.SubscribeInitialize(commandsForwardedSqlTopic)
	if err != nil {
		return nil, fmt.Errorf("error initialising command forwarder subscriber: %v", err)
	}

	cmdForwarderPublisher, err := NewRabbitMqPublisher(amqpURL, logger)
	if err != nil {
		return nil, err
	}

	cmdForwarder, err := forwarder.NewForwarder(cmdForwarderSubscriber, cmdForwarderPublisher, logger, forwarder.Config{
		ForwarderTopic: commandsForwardedSqlTopic,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating forwarder: %v", err)
	}

	cmdProcessor, err := m.newCommandProcessor()
	if err != nil {
		return nil, err
	}

	m.commandForwarder = cmdForwarder
	m.commandProcessor = cmdProcessor

	return m, nil
}

func (m *Messaging) Router() *message.Router {
	return m.router
}

func (m *Messaging) CommandForwarder() *forwarder.Forwarder {
	return m.commandForwarder
}

func (m *Messaging) CommandProcessor() *cqrs.CommandProcessor {
	return m.commandProcessor
}

func (m *Messaging) Close() error {
	return errors.Join(m.commandForwarder.Close(), m.router.Close())
}

func (m *Messaging) newCommandProcessor() (*cqrs.CommandProcessor, error) {
	return cqrs.NewCommandProcessorWithConfig(m.router, cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
			return "commands." + params.CommandName, nil
		},
		SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
			return amqp.NewSubscriber(amqp.NewDurableQueueConfig(m.amqpURL), m.logger)
		},
		OnHandle: nil,
		Marshaler: cqrs.JSONMarshaler{
			GenerateName: cqrs.StructName,
		},
		Logger:                   nil,
		AckCommandHandlingErrors: false,
	})
}

func (m *Messaging) registerMiddlewares(publisher message.Publisher) error {
	retry := middleware.Retry{
		MaxRetries:      3,
		InitialInterval: time.Millisecond * 300,
		Logger:          m.logger,
	}

	throttler := middleware.NewThrottle(maxMessagesPerSecond, time.Second)
	poisonQueue, err := middleware.PoisonQueue(publisher, poisonQueueName)
	if err != nil {
		return fmt.Errorf("error initialising the poison queue middleware: %v", err)
	}

	m.router.AddMiddleware(
		// Throttle messages
		throttler.Middleware,

		// Retry nacked messages
		retry.Middleware,

		//
		observabilityMiddleware(m.baseLogger),

		// Send nacked messages to a special queue
		poisonQueue,

		// Recover from panics
		middleware.Recoverer,
	)

	return nil
}

func (m *Messaging) Logger() watermill.LoggerAdapter {
	return m.logger
}
