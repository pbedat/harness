package board

import (
	"fmt"
	"maps"

	"github.com/samber/lo"
)

func (b *Board) AddColumn(columnName string) error {
	if columnName == "" {
		return fmt.Errorf("column name cannot be empty")
	}

	if _, exists := b.columns[columnName]; exists {
		return fmt.Errorf("column with name '%s' already exists", columnName)
	}

	b.columns[columnName] = &Column{
		name:  columnName,
		cards: make(map[string]*Card),
	}
	b.columnOrder = append(b.columnOrder, columnName)

	return nil
}

// RemoveColumn removes a column from the board and moves its cards to the nearest preceding column,
// or the next column if it is first. At least one column must remain on the board.
func (b *Board) RemoveColumn(columnName string) error {
	if len(b.columns) == 1 {
		return fmt.Errorf("cannot remove the only column in the board")
	}

	col, err := b.getColumn(columnName)
	if err != nil {
		return err
	}

	// Find the nearest preceding column; fall back to the next column if removing the first
	target := ""
	for i, name := range b.columnOrder {
		if name == columnName {
			if target == "" && i+1 < len(b.columnOrder) {
				target = b.columnOrder[i+1]
			}
			break
		}
		target = name
	}

	maps.Copy(b.columns[target].cards, col.cards)

	delete(b.columns, columnName)
	b.columnOrder = lo.Without(b.columnOrder, columnName)

	return nil
}
