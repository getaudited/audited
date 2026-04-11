package amqp

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/firminochangani/audited/internal/app"
)

func NewCommandHandlers(application *app.App) []cqrs.CommandHandler {
	return []cqrs.CommandHandler{}
}
