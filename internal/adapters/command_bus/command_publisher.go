package command_bus

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/tachyonhqdev/webhooks/internal/app/command"
)

type CommandSender struct {
	commandBus *cqrs.CommandBus
}

func NewCommandSender(commandBus *cqrs.CommandBus) CommandSender {
	return CommandSender{
		commandBus: commandBus,
	}
}

func (c CommandSender) SendCallEndpointCommand(ctx context.Context, cmd command.CallEndpointCommand) error {
	return c.commandBus.Send(ctx, cmd)
}
