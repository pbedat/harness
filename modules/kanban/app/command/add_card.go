package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type AddCard struct {
	BoardID     string
	Column      string
	ID          string
	Title       string
	Description string
	Assignee    *string
}

type AddCardHandler decorator.CommandHandler[AddCard]

func NewAddCardHandler(
	repo board.Repository,
	logger zerolog.Logger,
) AddCardHandler {
	return decorator.ApplyCommandDecorators(
		&addCardHandler{repo: repo}, logger,
	)
}

type addCardHandler struct {
	repo board.Repository
}

func (h addCardHandler) Handle(ctx context.Context, cmd AddCard) error {
	b, err := h.repo.Get(ctx, cmd.BoardID)
	if err != nil {
		return err
	}

	if err := b.AddCard(&board.AddCardDTO{
		ID:          cmd.ID,
		Title:       cmd.Title,
		Description: cmd.Description,
		Assignee:    cmd.Assignee,
	}, cmd.Column); err != nil {
		return err
	}

	return h.repo.Update(ctx, cmd.BoardID, b)
}
