package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type EditCard struct {
	BoardID  string
	ID       string
	Title    *string
	Body     *string
	Assignee *string
}

type EditCardHandler decorator.CommandHandler[EditCard]

func NewEditCardHandler(
	repo board.Repository,
	logger zerolog.Logger,
) EditCardHandler {
	return decorator.ApplyCommandDecorators(
		&editCardHandler{repo: repo}, logger,
	)
}

type editCardHandler struct {
	repo board.Repository
}

func (h editCardHandler) Handle(ctx context.Context, cmd EditCard) error {
	b, err := h.repo.Get(ctx, cmd.BoardID)
	if err != nil {
		return err
	}

	if err := b.EditCard(&board.EditCardDTO{
		ID:       cmd.ID,
		Title:    cmd.Title,
		Body:     cmd.Body,
		Assignee: cmd.Assignee,
	}); err != nil {
		return err
	}

	return h.repo.Update(ctx, cmd.BoardID, b)
}
