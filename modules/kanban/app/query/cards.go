package query

import (
	"context"
	"slices"
	"time"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/kanban/domain/board"
	"github.com/rs/zerolog"
)

type Cards struct {
	BoardID string
	Column  string
	Limit   int
}

type CardReadModel struct {
	ID          string
	Title       string
	Description string
	Assignee    *string
	ModifiedAt  time.Time
}

type CardsHandler decorator.QueryHandler[Cards, []*CardReadModel]

func NewCardsHandler(
	logger zerolog.Logger,
	repo board.Repository,
) CardsHandler {
	return decorator.ApplyQueryDecorators(
		&cardsHandler{repo}, logger,
	)
}

type cardsHandler struct {
	repo board.Repository
}

func (h cardsHandler) Handle(ctx context.Context, q Cards) ([]*CardReadModel, error) {
	board, err := h.repo.Get(ctx, q.BoardID)
	if err != nil {
		return nil, err
	}

	col, err := board.Column(q.Column)
	if err != nil {
		return nil, err
	}

	cards := slices.Collect(col.Cards())

	readModels := make([]*CardReadModel, 0, len(cards))
	for _, card := range cards {
		readModels = append(readModels, &CardReadModel{
			ID:          card.ID(),
			Title:       card.Title(),
			Description: card.Description(),
			Assignee:    card.Assignee(),
			ModifiedAt:  card.ModifiedAt(),
		})
	}

	return readModels, nil
}
