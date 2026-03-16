package query

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type Card struct {
	BoardID string
	CardID  string
}

type CardHandler decorator.QueryHandler[Card, *CardReadModel]

func NewCardHandler(logger zerolog.Logger, repo board.Repository) CardHandler {
	return decorator.ApplyQueryDecorators(
		&cardHandler{repo}, logger,
	)
}

type cardHandler struct {
	repo board.Repository
}

func (h cardHandler) Handle(ctx context.Context, q Card) (*CardReadModel, error) {
	b, err := h.repo.Get(ctx, q.BoardID)
	if err != nil {
		return nil, err
	}

	card, err := b.Card(q.CardID)
	if err != nil {
		return nil, err
	}

	return &CardReadModel{
		ID:          card.ID(),
		Title:       card.Title(),
		Description: card.Description(),
		Assignee:    card.Assignee(),
		ModifiedAt:  card.ModifiedAt(),
	}, nil
}
