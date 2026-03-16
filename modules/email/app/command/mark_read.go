package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/email/domain/email"
	"github.com/rs/zerolog"
)

type MarkRead struct {
	MailID string
}

type MarkReadHandler decorator.CommandHandler[MarkRead]

func NewMarkReadHandler(
	repo email.Repository,
	logger zerolog.Logger,
) MarkReadHandler {
	return decorator.ApplyCommandDecorators(
		&markReadHandler{repo: repo}, logger,
	)
}

type markReadHandler struct {
	repo email.Repository
}

func (h markReadHandler) Handle(ctx context.Context, cmd MarkRead) error {
	return h.repo.Update(ctx, cmd.MailID, func(e *email.Email) error {
		e.MarkAsRead()
		return nil
	})
}
