package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/queue"
	"github.com/rs/zerolog"
)

type Enqueue struct {
	Mailbox domain.Mailbox
	Mail    *queue.EmailDTO
}

type EnqueueHandler decorator.CommandHandler[Enqueue]

func NewEnqueueHandler(
	repo queue.Repository,
	logger zerolog.Logger,
) EnqueueHandler {
	return decorator.ApplyCommandDecorators(
		&enqueueHandler{repo: repo}, logger,
	)
}

type enqueueHandler struct {
	repo queue.Repository
}

func (h enqueueHandler) Handle(ctx context.Context, cmd Enqueue) error {
	return h.repo.Update(ctx, cmd.Mailbox, func(q *queue.Queue) error {
		mail, err := cmd.Mail.ToEmail()
		if err != nil {
			return err
		}
		return q.Enqueue(mail)
	})
}
