package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type RemoveColumn struct {
	BoardID    string
	ColumnName string
}

type RemoveColumnHandler decorator.CommandHandler[RemoveColumn]

func NewRemoveColumnHandler(
	repo board.Repository,
	logger zerolog.Logger,
) RemoveColumnHandler {
	return decorator.ApplyCommandDecorators(
		&removeColumnHandler{repo: repo}, logger,
	)
}

type removeColumnHandler struct {
	repo board.Repository
}

func (h removeColumnHandler) Handle(ctx context.Context, cmd RemoveColumn) error {
	b, err := h.repo.Get(ctx, cmd.BoardID)
	if err != nil {
		return err
	}

	if err := b.RemoveColumn(cmd.ColumnName); err != nil {
		return err
	}

	return h.repo.Update(ctx, cmd.BoardID, b)
}
