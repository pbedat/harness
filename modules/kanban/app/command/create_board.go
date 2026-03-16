package command

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type CreateBoard struct {
	ID      string
	Name    string
	Columns []string
}

type CreateBoardHandler decorator.CommandHandler[CreateBoard]

func NewCreateBoardHandler(
	repo board.Repository,
	logger zerolog.Logger,
) CreateBoardHandler {
	return decorator.ApplyCommandDecorators(
		&createBoardHandler{repo: repo}, logger,
	)
}

type createBoardHandler struct {
	repo board.Repository
}

func (h createBoardHandler) Handle(ctx context.Context, cmd CreateBoard) error {
	b, err := board.NewBoard(cmd.ID, cmd.Name, cmd.Columns...)
	if err != nil {
		return err
	}
	return h.repo.Create(ctx, b)
}
