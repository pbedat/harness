package command

import (
	"context"
	"time"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/domain/board"
	"github.com/rs/zerolog"
)

type ArchiveCards struct {
	BoardID       string
	Column        string
	StaleDuration time.Duration
}

type ArchiveCardsHandler decorator.CommandHandler[ArchiveCards]

func NewArchiveCardsHandler(
	repo board.Repository,
	logger zerolog.Logger,
) ArchiveCardsHandler {
	return decorator.ApplyCommandDecorators(
		&archiveCardsHandler{repo: repo}, logger,
	)
}

type archiveCardsHandler struct {
	repo board.Repository
}

func (h archiveCardsHandler) Handle(ctx context.Context, cmd ArchiveCards) error {
	b, err := h.repo.Get(ctx, cmd.BoardID)
	if err != nil {
		return err
	}

	if err := b.ArchiveCards(cmd.Column, cmd.StaleDuration); err != nil {
		return err
	}

	return h.repo.Update(ctx, cmd.BoardID, b)
}
