package amqp

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/tachyonhqdev/webhooks/internal/app"
)

func NewCommandHandlers(application *app.App) []cqrs.CommandHandler {
	return []cqrs.CommandHandler{}
}
