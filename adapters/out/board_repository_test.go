package out

import (
	"context"
	"embed"
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/pbedat/harness/domain/board"
	"github.com/spf13/afero"
)

//go:embed testdata/boards
var testBoardsFS embed.FS

func TestBoardFSRepoGet(t *testing.T) {

	memFS := newMemFS()

	sut := NewBoardFSRepository(memFS, "testdata/boards")

	t.Run("columns", func(t *testing.T) {

		board, err := sut.Get(t.Context(), "1")
		if err != nil {
			t.Fatal(err)
		}

		actual := []string{}

		for col := range board.Columns() {
			actual = append(actual, col.Name())
		}

		expected := []string{"To Do", "In Progress", "Done"}

		if !slices.Equal(expected, actual) {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	})

	t.Run("cards", func(t *testing.T) {
		brd, err := sut.Get(context.Background(), "1")
		if err != nil {
			t.Fatal(err)
		}

		card := brd.FindCard("ac4e4034-0d20-41c6-9da4-24afd5f9e80e")
		if card == nil {
			t.Fatal("expected card ac4e4034 to be present")
		}
		if card.Title() != "Implement the Repositories" {
			t.Errorf("unexpected title: %q", card.Title())
		}
		if card.Assignee() == nil || *card.Assignee() != "@pbedat" {
			t.Errorf("expected assignee '@pbedat', got %v", *card.Assignee())
		}
		expectedTime := time.Date(2026, 2, 23, 12, 37, 0, 0, time.UTC)
		if !card.ModifiedAt().Equal(expectedTime) {
			t.Errorf("expected modifiedAt %v, got %v", expectedTime, card.ModifiedAt())
		}
	})
}

func TestBoardFSRepoGetNotFound(t *testing.T) {
	memFS := afero.NewMemMapFs()
	_ = memFS.MkdirAll("boards", 0755)
	sut := NewBoardFSRepository(memFS, "boards")

	_, err := sut.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent board")
	}
}

func TestBoardFSRepoCreate(t *testing.T) {
	memFS := afero.NewMemMapFs()
	sut := NewBoardFSRepository(memFS, "boards")

	brd, err := board.NewBoard("42", "New Board", "Backlog", "Done")
	if err != nil {
		t.Fatal(err)
	}

	if err := sut.Create(context.Background(), brd); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := sut.Get(context.Background(), "42")
	if err != nil {
		t.Fatalf("Get after Create failed: %v", err)
	}
	if got.ID() != "42" {
		t.Errorf("expected id '42', got %q", got.ID())
	}
	if got.Name() != "New Board" {
		t.Errorf("expected name 'New Board', got %q", got.Name())
	}

	var cols []string
	for c := range got.Columns() {
		cols = append(cols, c.Name())
	}
	if !slices.Equal([]string{"Backlog", "Done"}, cols) {
		t.Errorf("unexpected columns: %v", cols)
	}
}

func TestBoardFSRepoCreateWithCards(t *testing.T) {
	memFS := afero.NewMemMapFs()
	sut := NewBoardFSRepository(memFS, "boards")

	brd, _ := board.NewBoard("10", "Board With Cards", "To Do", "Done")
	assignee := "alice"
	_ = brd.AddCard(&board.AddCardDTO{
		ID:          "card-1",
		Title:       "First Card",
		Description: "Some description",
		Assignee:    &assignee,
	}, "To Do")
	_ = brd.AddCard(&board.AddCardDTO{
		ID:    "card-2",
		Title: "Second Card",
	}, "Done")

	if err := sut.Create(context.Background(), brd); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	got, err := sut.Get(context.Background(), "10")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	card := got.FindCard("card-1")
	if card == nil {
		t.Fatal("card-1 not found after round-trip")
	}
	if card.Title() != "First Card" {
		t.Errorf("expected title 'First Card', got %q", card.Title())
	}
	if card.Description() != "Some description" {
		t.Errorf("expected description 'Some description', got %q", card.Description())
	}
	if card.Assignee() == nil || *card.Assignee() != "alice" {
		t.Errorf("expected assignee 'alice', got %v", *card.Assignee())
	}
	if got.FindCard("card-2") == nil {
		t.Fatal("card-2 not found after round-trip")
	}
}

func TestBoardFSRepoUpdate(t *testing.T) {
	dir := t.TempDir()
	sut := NewBoardFSRepository(afero.NewOsFs(), dir)

	brd, _ := board.NewBoard("5", "Original", "To Do")
	_ = sut.Create(context.Background(), brd)

	_ = brd.AddCard(&board.AddCardDTO{ID: "c1", Title: "Updated Card"}, "To Do")
	if err := sut.Update(context.Background(), "5", brd); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, err := sut.Get(context.Background(), "5")
	if err != nil {
		t.Fatalf("Get after Update failed: %v", err)
	}
	if got.FindCard("c1") == nil {
		t.Error("expected card c1 after update")
	}
}

func TestBoardFSRepoUpdateBlocks(t *testing.T) {
	dir := t.TempDir()
	sut := NewBoardFSRepository(afero.NewOsFs(), dir)

	brd, _ := board.NewBoard("99", "Lock Test", "To Do")
	_ = sut.Create(context.Background(), brd)

	// Hold an exclusive flock on the lock file, release after a delay.
	lf, _ := os.OpenFile(filepath.Join(dir, "99.lock"), os.O_CREATE|os.O_WRONLY, 0600)
	_ = syscall.Flock(int(lf.Fd()), syscall.LOCK_EX)
	go func() {
		time.Sleep(50 * time.Millisecond)
		syscall.Flock(int(lf.Fd()), syscall.LOCK_UN)
		lf.Close()
	}()

	if err := sut.Update(context.Background(), "99", brd); err != nil {
		t.Fatalf("expected Update to succeed after lock released, got: %v", err)
	}
}

func TestBoardFSRepoUpdateContextCancelled(t *testing.T) {
	dir := t.TempDir()
	sut := NewBoardFSRepository(afero.NewOsFs(), dir)

	brd, _ := board.NewBoard("99", "Lock Test", "To Do")
	_ = sut.Create(context.Background(), brd)

	// Hold an exclusive flock permanently.
	lf, _ := os.OpenFile(filepath.Join(dir, "99.lock"), os.O_CREATE|os.O_WRONLY, 0600)
	_ = syscall.Flock(int(lf.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(lf.Fd()), syscall.LOCK_UN)
	defer lf.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	if err := sut.Update(ctx, "99", brd); err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}


func TestParseMarkdown(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedTitle   string
		expectedHeaders map[string]string
		expectedBody    string
	}{
		{
			name:            "empty",
			input:           "",
			expectedTitle:   "",
			expectedHeaders: map[string]string{},
			expectedBody:    "",
		},
		{
			name:            "title only",
			input:           "# My Title\n",
			expectedTitle:   "My Title",
			expectedHeaders: map[string]string{},
			expectedBody:    "",
		},
		{
			name:            "full card",
			input:           "# Card Title\n\nAssignee: @alice\nID: card-1\n\n---\n\nCard body here\n",
			expectedTitle:   "Card Title",
			expectedHeaders: map[string]string{"Assignee": "@alice", "ID": "card-1"},
			expectedBody:    "Card body here",
		},
		{
			name:            "no separator",
			input:           "# Title\n\nKey: Value\n",
			expectedTitle:   "Title",
			expectedHeaders: map[string]string{"Key": "Value"},
			expectedBody:    "",
		},
		{
			name:            "multiline body",
			input:           "# Title\n\n---\n\nLine 1\nLine 2\n",
			expectedTitle:   "Title",
			expectedHeaders: map[string]string{},
			expectedBody:    "Line 1\nLine 2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			title, headers, body := parseMarkdown(tc.input)
			if title != tc.expectedTitle {
				t.Errorf("title: expected %q, got %q", tc.expectedTitle, title)
			}
			if body != tc.expectedBody {
				t.Errorf("body: expected %q, got %q", tc.expectedBody, body)
			}
			for k, v := range tc.expectedHeaders {
				if headers[k] != v {
					t.Errorf("header %q: expected %q, got %q", k, v, headers[k])
				}
			}
		})
	}
}

func TestMarshalCard(t *testing.T) {
	t.Run("without assignee", func(t *testing.T) {
		c, _ := board.NewCard("c1", "My Card")
		c.SetDescription("Some body")
		out := marshalCard(c)
		if !strings.Contains(out, "# My Card") {
			t.Error("expected title in output")
		}
		if !strings.Contains(out, "ID: c1") {
			t.Error("expected ID in output")
		}
		if !strings.Contains(out, "Some body") {
			t.Error("expected description in output")
		}
		if strings.Contains(out, "Assignee:") {
			t.Error("expected no Assignee line when nil")
		}
	})

	t.Run("with assignee", func(t *testing.T) {
		c, _ := board.NewCard("c2", "Another Card")
		_ = c.SetAssignee("bob")
		out := marshalCard(c)
		if !strings.Contains(out, "Assignee: bob") {
			t.Errorf("expected 'Assignee: bob' in output, got:\n%s", out)
		}
	})
}

func TestSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"To Do", "to-do"},
		{"In Progress", "in-progress"},
		{"Hello World", "hello-world"},
		{"ALL CAPS", "all-caps"},
		{"123 Board", "123-board"},
	}
	for _, tc := range tests {
		got := slug(tc.input)
		if got != tc.expected {
			t.Errorf("slug(%q): expected %q, got %q", tc.input, tc.expected, got)
		}
	}
}

func TestNumericPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		ok       bool
	}{
		{"1_todo", 1, true},
		{"42_done", 42, true},
		{"no-underscore", 0, false},
		{"abc_col", 0, false},
	}
	for _, tc := range tests {
		n, ok := numericPrefix(tc.input)
		if ok != tc.ok {
			t.Errorf("numericPrefix(%q): expected ok=%v, got %v", tc.input, tc.ok, ok)
		}
		if ok && n != tc.expected {
			t.Errorf("numericPrefix(%q): expected %d, got %d", tc.input, tc.expected, n)
		}
	}
}

func newMemFS() afero.Fs {
	memFS := afero.NewMemMapFs()

	err := afero.Walk(afero.FromIOFS{FS: testBoardsFS}, "testdata/boards", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return memFS.MkdirAll(path, info.Mode())
		}
		data, err := testBoardsFS.ReadFile(path)
		if err != nil {
			return err
		}
		return afero.WriteFile(memFS, path, data, info.Mode())
	})

	if err != nil {
		panic(err)
	}

	return memFS
}
