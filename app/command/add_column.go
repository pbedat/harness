package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/domain/board"
	"github.com/rs/zerolog"
)

type AddColumn struct {
	BoardID    string
	ColumnName string
}

type AddColumnHandler decorator.CommandHandler[AddColumn]

func NewAddColumnHandler(
	repo board.Repository,
	logger zerolog.Logger,
) AddColumnHandler {
	return decorator.ApplyCommandDecorators(
		&addColumnHandler{repo: repo}, logger,
	)
}

type addColumnHandler struct {
	repo board.Repository
}

func (h addColumnHandler) Handle(ctx context.Context, cmd AddColumn) error {
	b, err := h.repo.Get(ctx, cmd.BoardID)
	if err != nil {
		return err
	}

	if err := b.AddColumn(cmd.ColumnName); err != nil {
		return err
	}

	return h.repo.Update(ctx, cmd.BoardID, b)
}
