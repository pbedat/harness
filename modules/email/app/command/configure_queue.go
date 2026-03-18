package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/rs/zerolog"
)

type ConfigureQueue struct {
	Mailbox           domain.Mailbox
	AllowedRecipients []string
	AllowedFrom       *string
}

type QueueConfigWriter interface {
	WriteQueueMeta(mailbox domain.Mailbox, allowedRecipients []string, allowedFrom *string) error
}

type ConfigureQueueHandler decorator.CommandHandler[ConfigureQueue]

func NewConfigureQueueHandler(
	writer QueueConfigWriter,
	logger zerolog.Logger,
) ConfigureQueueHandler {
	return decorator.ApplyCommandDecorators(
		&configureQueueHandler{writer: writer}, logger,
	)
}

type configureQueueHandler struct {
	writer QueueConfigWriter
}

func (h configureQueueHandler) Handle(ctx context.Context, cmd ConfigureQueue) error {
	return h.writer.WriteQueueMeta(cmd.Mailbox, cmd.AllowedRecipients, cmd.AllowedFrom)
}
