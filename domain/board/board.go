package board

import (
	"fmt"
	"iter"
	"time"
)

func NewBoard(id, name string, columns ...string) (*Board, error) {

	if len(columns) == 0 {
		return nil, fmt.Errorf("board needs at least one column")
	}

	if id == "" {
		return nil, fmt.Errorf("board id cannot be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("board name cannot be empty")
	}

	board := &Board{
		id:          id,
		name:        name,
		columns:     make(map[string]*Column, len(columns)),
		columnOrder: make([]string, 0, len(columns)),
	}

	for _, columnName := range columns {
		board.columns[columnName] = &Column{
			name:  columnName,
			cards: make(map[string]*Card),
		}
		board.columnOrder = append(board.columnOrder, columnName)
	}

	return board, nil
}

type Board struct {
	id          string
	name        string
	columns     map[string]*Column
	columnOrder []string
}

type Column struct {
	name string

	cards map[string]*Card
}

func (c *Column) Name() string {
	return c.name
}

func (c *Column) Cards() iter.Seq[*Card] {
	return func(yield func(*Card) bool) {
		for _, card := range c.cards {
			if !yield(card) {
				return
			}
		}
	}
}

type Card struct {
	id          string
	title       string
	description string
	assignee    *string
	modifiedAt  time.Time
}

func (b *Board) ID() string {
	return b.id
}

func (b *Board) Name() string {
	return b.name
}

func (b *Board) Column(name string) (*Column, error) {
	col, ok := b.columns[name]
	if !ok {
		return nil, fmt.Errorf("column with name %s not found", name)
	}
	return col, nil
}

type UnmarshalCardDTO struct {
	ID          string
	Title       string
	Description string
	Assignee    *string
	ModifiedAt  time.Time
}

type UnmarshalColumnDTO struct {
	Name  string
	Cards []*UnmarshalCardDTO
}

type UnmarshalBoardDTO struct {
	ID      string
	Name    string
	Columns []*UnmarshalColumnDTO
}

func UnmarshalBoard(dto *UnmarshalBoardDTO) *Board {
	b := &Board{
		id:          dto.ID,
		name:        dto.Name,
		columns:     make(map[string]*Column, len(dto.Columns)),
		columnOrder: make([]string, 0, len(dto.Columns)),
	}
	for _, colDTO := range dto.Columns {
		cards := make(map[string]*Card, len(colDTO.Cards))
		for _, c := range colDTO.Cards {
			cards[c.ID] = &Card{
				id:          c.ID,
				title:       c.Title,
				description: c.Description,
				assignee:    c.Assignee,
				modifiedAt:  c.ModifiedAt,
			}
		}
		b.columns[colDTO.Name] = &Column{name: colDTO.Name, cards: cards}
		b.columnOrder = append(b.columnOrder, colDTO.Name)
	}
	return b
}

func (b *Board) Card(id string) (*Card, error) {
	for _, col := range b.columns {
		if card, ok := col.cards[id]; ok {
			return card, nil
		}
	}
	return nil, fmt.Errorf("card with id %s not found", id)
}

func (b *Board) Columns() iter.Seq[*Column] {
	return func(yield func(*Column) bool) {
		for _, columnName := range b.columnOrder {
			if !yield(b.columns[columnName]) {
				return
			}
		}
	}
}
