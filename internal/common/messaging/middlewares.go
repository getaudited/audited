package messaging

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/tachyonhqdev/webhooks/internal/common/logs"
)

func observabilityMiddleware(logger *logs.Logger) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger.Info("Handling message", "id", msg.UUID)
			messages, err := h(msg)
			logger.Info("Message handled", "id", msg.UUID, "err", err)
			if err != nil {
				logger.Error("Error handling message", "id", msg.UUID, "error", err)
			}

			return messages, err
		}
	}
}
