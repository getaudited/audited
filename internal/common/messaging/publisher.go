package messaging

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type RabbitMqPublisher struct {
	publisher message.Publisher
}

func (p RabbitMqPublisher) Publish(topic string, messages ...*message.Message) error {
	return p.publisher.Publish(topic, messages...)
}

func (p RabbitMqPublisher) Close() error {
	return p.publisher.Close()
}

func NewRabbitMqPublisher(amqpURL string, logger watermill.LoggerAdapter) (message.Publisher, error) {
	publisher, err := amqp.NewPublisher(amqp.NewDurableQueueConfig(amqpURL), logger)
	if err != nil {
		return nil, err
	}

	return RabbitMqPublisher{
		publisher: publisher,
	}, nil
}
