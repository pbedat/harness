package board

import (
	"fmt"
	"time"
)

func NewCard(id, title string) (*Card, error) {
	if id == "" {
		return nil, fmt.Errorf("card id cannot be empty")
	}

	if title == "" {
		return nil, fmt.Errorf("card title cannot be empty")
	}

	return &Card{
		id:         id,
		title:      title,
		modifiedAt: time.Now(),
	}, nil
}

func (c *Card) SetDescription(description string) {
	c.description = description
	c.modifiedAt = time.Now()
}

func (c *Card) SetAssignee(assignee string) error {
	if assignee == "" {
		return fmt.Errorf("assignee cannot be empty")
	}
	c.assignee = &assignee
	c.modifiedAt = time.Now()
	return nil
}

func (b *Board) getColumn(columnName string) (*Column, error) {
	col, exists := b.columns[columnName]
	if !exists {
		return nil, fmt.Errorf("column with name '%s' does not exist", columnName)
	}
	return col, nil
}

func (b *Board) findCardColumn(cardID string) (*Column, *Card) {
	for _, col := range b.columns {
		if card, exists := col.cards[cardID]; exists {
			return col, card
		}
	}
	return nil, nil
}

type AddCardDTO struct {
	ID          string
	Title       string
	Description string
	Assignee    *string
}

func (b *Board) AddCard(dto *AddCardDTO, columnName string) error {
	card, err := NewCard(dto.ID, dto.Title)
	if err != nil {
		return err
	}

	card.SetDescription(dto.Description)
	if dto.Assignee != nil {
		if err := card.SetAssignee(*dto.Assignee); err != nil {
			return err
		}
	}

	column, err := b.getColumn(columnName)
	if err != nil {
		return err
	}

	if _, exists := column.cards[card.id]; exists {
		return fmt.Errorf("card with id '%s' already exists in column '%s'", card.id, columnName)
	}

	column.cards[card.id] = card

	return nil
}

func (b *Board) MoveCard(cardID, column string) error {
	sourceCol, card := b.findCardColumn(cardID)
	if sourceCol == nil {
		return fmt.Errorf("card with id '%s' not found in any column", cardID)
	}

	targetCol, err := b.getColumn(column)
	if err != nil {
		return err
	}

	delete(sourceCol.cards, cardID)
	targetCol.cards[cardID] = card
	card.modifiedAt = time.Now()

	return nil
}

func (b *Board) ArchiveCards(column string, stale time.Duration) error {
	col, err := b.getColumn(column)
	if err != nil {
		return err
	}

	now := time.Now()
	for id, card := range col.cards {
		if card.modifiedAt.Add(stale).Before(now) {
			delete(col.cards, id)
		}
	}

	// TODO: event here, to capture archived cards

	return nil
}

func (c *Card) ID() string {
	return c.id
}

func (c *Card) Title() string {
	return c.title
}

func (c *Card) Description() string {
	return c.description
}

func (c *Card) Assignee() *string {
	return c.assignee
}

func (c *Card) ModifiedAt() time.Time {
	return c.modifiedAt
}

func (c *Card) UnmarshalModifiedAt(t time.Time) {
	c.modifiedAt = t
}

func (b *Board) FindCard(id string) *Card {
	_, card := b.findCardColumn(id)
	return card
}

type EditCardDTO struct {
	ID       string
	Title    *string
	Body     *string
	Assignee *string
}

func (b *Board) EditCard(dto *EditCardDTO) error {
	_, existing := b.findCardColumn(dto.ID)
	if existing == nil {
		return fmt.Errorf("card with id '%s' not found in any column", dto.ID)
	}

	if dto.Title != nil {
		existing.title = *dto.Title
	}
	if dto.Body != nil {
		existing.description = *dto.Body
	}
	if dto.Assignee != nil {
		existing.assignee = dto.Assignee
	}
	existing.modifiedAt = time.Now()

	return nil
}
