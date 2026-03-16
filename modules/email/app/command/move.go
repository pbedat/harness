package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/email"
	"github.com/rs/zerolog"
)

type Move struct {
	MailID string
	To     domain.Mailbox
}

type MoveHandler decorator.CommandHandler[Move]

func NewMoveHandler(
	repo email.Repository,
	logger zerolog.Logger,
) MoveHandler {
	return decorator.ApplyCommandDecorators(
		&moveHandler{repo: repo}, logger,
	)
}

type moveHandler struct {
	repo email.Repository
}

func (h moveHandler) Handle(ctx context.Context, cmd Move) error {
	return h.repo.Update(ctx, cmd.MailID, func(e *email.Email) error {
		return e.Move(cmd.To)
	})
}
