package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type MoveCard struct {
	BoardID string
	CardID  string
	Column  string
}

type MoveCardHandler decorator.CommandHandler[MoveCard]

func NewMoveCardHandler(
	repo board.Repository,
	logger zerolog.Logger,
) MoveCardHandler {
	return decorator.ApplyCommandDecorators(
		&moveCardHandler{repo: repo}, logger,
	)
}

type moveCardHandler struct {
	repo board.Repository
}

func (h moveCardHandler) Handle(ctx context.Context, cmd MoveCard) error {
	b, err := h.repo.Get(ctx, cmd.BoardID)
	if err != nil {
		return err
	}

	if err := b.MoveCard(cmd.CardID, cmd.Column); err != nil {
		return err
	}

	return h.repo.Update(ctx, cmd.BoardID, b)
}
