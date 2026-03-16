package board

import (
	"slices"
	"testing"
	"time"
)

func TestNewBoard(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		b, err := NewBoard("1", "My Board", "To Do", "Done")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.ID() != "1" {
			t.Errorf("expected id '1', got %q", b.ID())
		}
		if b.Name() != "My Board" {
			t.Errorf("expected name 'My Board', got %q", b.Name())
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := NewBoard("", "My Board", "To Do")
		if err == nil {
			t.Fatal("expected error for empty id")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewBoard("1", "", "To Do")
		if err == nil {
			t.Fatal("expected error for empty name")
		}
	})

	t.Run("no columns", func(t *testing.T) {
		_, err := NewBoard("1", "My Board")
		if err == nil {
			t.Fatal("expected error for no columns")
		}
	})
}

func TestBoardColumns(t *testing.T) {
	b, _ := NewBoard("1", "Board", "To Do", "In Progress", "Done")

	var names []string
	for col := range b.Columns() {
		names = append(names, col.Name())
	}
	expected := []string{"To Do", "In Progress", "Done"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(names))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("column %d: expected %q, got %q", i, name, names[i])
		}
	}
}

func TestNewCard(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c, err := NewCard("c1", "My Card")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID() != "c1" {
			t.Errorf("expected id 'c1', got %q", c.ID())
		}
		if c.Title() != "My Card" {
			t.Errorf("expected title 'My Card', got %q", c.Title())
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := NewCard("", "My Card")
		if err == nil {
			t.Fatal("expected error for empty id")
		}
	})

	t.Run("empty title", func(t *testing.T) {
		_, err := NewCard("c1", "")
		if err == nil {
			t.Fatal("expected error for empty title")
		}
	})
}

func TestCardSetDescription(t *testing.T) {
	c, _ := NewCard("c1", "Title")
	before := c.ModifiedAt()
	time.Sleep(time.Millisecond)
	c.SetDescription("my desc")
	if c.Description() != "my desc" {
		t.Errorf("expected description 'my desc', got %q", c.Description())
	}
	if !c.ModifiedAt().After(before) {
		t.Error("expected modifiedAt to be updated after SetDescription")
	}
}

func TestCardSetAssignee(t *testing.T) {
	c, _ := NewCard("c1", "Title")

	t.Run("valid", func(t *testing.T) {
		before := c.ModifiedAt()
		time.Sleep(time.Millisecond)
		err := c.SetAssignee("alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Assignee() == nil || *c.Assignee() != "alice" {
			t.Errorf("expected assignee 'alice', got %v", c.Assignee())
		}
		if !c.ModifiedAt().After(before) {
			t.Error("expected modifiedAt to be updated after SetAssignee")
		}
	})

	t.Run("empty assignee", func(t *testing.T) {
		err := c.SetAssignee("")
		if err == nil {
			t.Fatal("expected error for empty assignee")
		}
	})
}

func TestCardUnmarshalModifiedAt(t *testing.T) {
	c, _ := NewCard("c1", "Title")
	ts := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c.UnmarshalModifiedAt(ts)
	if !c.ModifiedAt().Equal(ts) {
		t.Errorf("expected modifiedAt %v, got %v", ts, c.ModifiedAt())
	}
}

func TestAddCard(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.AddCard(&AddCardDTO{ID: "c1", Title: "Card"}, "To Do")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unknown column", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.AddCard(&AddCardDTO{ID: "c1", Title: "Card"}, "Nonexistent")
		if err == nil {
			t.Fatal("expected error for unknown column")
		}
	})

	t.Run("duplicate card id", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Card"}, "To Do")
		err := b.AddCard(&AddCardDTO{ID: "c1", Title: "Another"}, "To Do")
		if err == nil {
			t.Fatal("expected error for duplicate card id")
		}
	})

	t.Run("with assignee", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		assignee := "alice"
		err := b.AddCard(&AddCardDTO{ID: "c1", Title: "Card", Assignee: &assignee}, "To Do")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		card := b.FindCard("c1")
		if card == nil || card.Assignee() == nil || *card.Assignee() != "alice" {
			t.Error("expected assignee 'alice'")
		}
	})

	t.Run("invalid card id", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.AddCard(&AddCardDTO{ID: "", Title: "Card"}, "To Do")
		if err == nil {
			t.Fatal("expected error for empty card id")
		}
	})
}

func TestMoveCard(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do", "Done")
		_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Card"}, "To Do")
		err := b.MoveCard("c1", "Done")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.FindCard("c1") == nil {
			t.Fatal("card not found after move")
		}
	})

	t.Run("card not found", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do", "Done")
		err := b.MoveCard("nonexistent", "Done")
		if err == nil {
			t.Fatal("expected error for missing card")
		}
	})

	t.Run("target column not found", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Card"}, "To Do")
		err := b.MoveCard("c1", "Nonexistent")
		if err == nil {
			t.Fatal("expected error for unknown column")
		}
	})
}

func TestArchiveCards(t *testing.T) {
	t.Run("archives stale cards", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "Done")
		_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Old Card"}, "Done")
		// Set modifiedAt to the past
		b.FindCard("c1").UnmarshalModifiedAt(time.Now().Add(-48 * time.Hour))

		err := b.ArchiveCards("Done", 24*time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.FindCard("c1") != nil {
			t.Error("expected stale card to be archived")
		}
	})

	t.Run("keeps fresh cards", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "Done")
		_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Fresh Card"}, "Done")

		err := b.ArchiveCards("Done", 24*time.Hour)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if b.FindCard("c1") == nil {
			t.Error("expected fresh card to remain")
		}
	})

	t.Run("unknown column", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "Done")
		err := b.ArchiveCards("Nonexistent", time.Hour)
		if err == nil {
			t.Fatal("expected error for unknown column")
		}
	})
}

func TestFindCard(t *testing.T) {
	b, _ := NewBoard("1", "Board", "To Do")
	_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Card"}, "To Do")

	if b.FindCard("c1") == nil {
		t.Error("expected to find card c1")
	}
	if b.FindCard("nonexistent") != nil {
		t.Error("expected nil for nonexistent card")
	}
}

func TestEditCard(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Old Title"}, "To Do")
		assignee := "bob"
		err := b.EditCard(&EditCardDTO{ID: "c1", Title: "New Title", Body: "New Body", Assignee: &assignee})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		card := b.FindCard("c1")
		if card.Title() != "New Title" {
			t.Errorf("expected title 'New Title', got %q", card.Title())
		}
		if card.Description() != "New Body" {
			t.Errorf("expected description 'New Body', got %q", card.Description())
		}
		if card.Assignee() == nil || *card.Assignee() != "bob" {
			t.Error("expected assignee 'bob'")
		}
	})

	t.Run("card not found", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.EditCard(&EditCardDTO{ID: "nonexistent", Title: "Title"})
		if err == nil {
			t.Fatal("expected error for nonexistent card")
		}
	})
}

func TestColumnCards(t *testing.T) {
	b, _ := NewBoard("1", "Board", "To Do")
	_ = b.AddCard(&AddCardDTO{ID: "c1", Title: "Card 1"}, "To Do")
	_ = b.AddCard(&AddCardDTO{ID: "c2", Title: "Card 2"}, "To Do")

	var col *Column
	for c := range b.Columns() {
		col = c
	}
	cards := slices.Collect(col.Cards())
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
}

func TestUnmarshalBoard(t *testing.T) {
	assignee := "alice"
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	b := UnmarshalBoard(&UnmarshalBoardDTO{
		ID:   "u1",
		Name: "Unmarshalled Board",
		Columns: []*UnmarshalColumnDTO{
			{
				Name: "To Do",
				Cards: []*UnmarshalCardDTO{
					{ID: "c1", Title: "Card 1", Description: "desc", Assignee: &assignee, ModifiedAt: ts},
				},
			},
			{Name: "Done", Cards: nil},
		},
	})

	if b.ID() != "u1" {
		t.Errorf("expected id 'u1', got %q", b.ID())
	}
	if b.Name() != "Unmarshalled Board" {
		t.Errorf("expected name 'Unmarshalled Board', got %q", b.Name())
	}

	card := b.FindCard("c1")
	if card == nil {
		t.Fatal("expected card c1")
	}
	if card.Title() != "Card 1" || card.Description() != "desc" {
		t.Errorf("card fields mismatch: %q %q", card.Title(), card.Description())
	}
	if card.Assignee() == nil || *card.Assignee() != "alice" {
		t.Errorf("expected assignee alice, got %v", card.Assignee())
	}
	if !card.ModifiedAt().Equal(ts) {
		t.Errorf("expected modifiedAt %v, got %v", ts, card.ModifiedAt())
	}
}

func TestColumnsEarlyExit(t *testing.T) {
	b, _ := NewBoard("1", "Board", "A", "B", "C")
	count := 0
	for range b.Columns() {
		count++
		break
	}
	if count != 1 {
		t.Errorf("expected early exit after 1 column, got %d", count)
	}
}

func TestAddColumn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.AddColumn("Done")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.AddColumn("")
		if err == nil {
			t.Fatal("expected error for empty column name")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.AddColumn("To Do")
		if err == nil {
			t.Fatal("expected error for duplicate column name")
		}
	})
}

func TestRemoveColumn(t *testing.T) {
	t.Run("only column", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do")
		err := b.RemoveColumn("To Do")
		if err == nil {
			t.Fatal("expected error when removing only column")
		}
	})

	t.Run("nonexistent column", func(t *testing.T) {
		b, _ := NewBoard("1", "Board", "To Do", "Done")
		err := b.RemoveColumn("Nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent column")
		}
	})
}
