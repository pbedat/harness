package board

import (
	"testing"

	"github.com/samber/lo"
)

func TestColumns(t *testing.T) {

	t.Run("remove", func(t *testing.T) {
		board, err := NewBoard("1", "Test Board", "To Do", "In Progress", "Done")
		if err != nil {
			t.Fatalf("unexpected error creating board: %v", err)
		}

		c1, _ := NewCard("1", "Test Card")
		c2, _ := NewCard("2", "Test Card 2")

		lo.Must0(board.AddCard(&AddCardDTO{ID: c1.id, Title: c1.title}, "Done"))
		lo.Must0(board.AddCard(&AddCardDTO{ID: c2.id, Title: c2.title}, "Done"))

		if err := board.RemoveColumn("Done"); err != nil {
			t.Fatalf("unexpected error removing column: %v", err)
		}

		cards := board.columns["In Progress"].cards

		if len(cards) != 2 {
			t.Fatal("expected cards to be moved to 'In Progress' column")
		}
	})

	t.Run("remove first column", func(t *testing.T) {
		board, err := NewBoard("1", "Test Board", "To Do", "In Progress", "Done")
		if err != nil {
			t.Fatalf("unexpected error creating board: %v", err)
		}

		c1, _ := NewCard("1", "Test Card")
		c2, _ := NewCard("2", "Test Card 2")

		lo.Must0(board.AddCard(&AddCardDTO{ID: c1.id, Title: c1.title}, "To Do"))
		lo.Must0(board.AddCard(&AddCardDTO{ID: c2.id, Title: c2.title}, "To Do"))

		if err := board.RemoveColumn("To Do"); err != nil {
			t.Fatalf("unexpected error removing column: %v", err)
		}

		cards := board.columns["In Progress"].cards

		if len(cards) != 2 {
			t.Fatal("expected cards to be moved to 'In Progress' column")
		}
	})
}
