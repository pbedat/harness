package out

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/pbedat/harness/domain/board"
	"github.com/spf13/afero"
)

const (
	boardFile  = "board.md"
	columnFile = "column.md"
)

// BoardFSRepository implements the [board.Repository] interface
// using the file system for storage.
// Every board is stored in a separate folder named after the board
// Columns are stored as separate folders underneath the board folder
// Each card is a markdown file with a fixed layout in the column folder
//
// Example:
// boards/
//
//			board-1/
//		   board.md (contains board metadata)
//			  1_to-do/
//	        column.md (contains column metadata)
//			    1_create-repository.md
//			    2_add-commands.md
//			  2_in-progress/
//	        column.md (contains column metadata)
//			    3_implement-board.md
//			  3_done/
//					column.md (contains column metadata)
//			    4_setup-project.md
//
// Example card:
//
//	# Implement Board
//	Assignee: @pbedat
//	Modified At: 2024-06-01T12:00:00Z
//	---------------------------------
//	Implement the Board domain package
type BoardFSRepository struct {
	fs       afero.Fs
	basePath string
}

// Create implements [board.Repository].
func (b *BoardFSRepository) Create(ctx context.Context, brd *board.Board) error {
	boardDir := filepath.Join(b.basePath, brd.ID()+"_"+slug(brd.Name()))
	return b.writeBoardToDisk(brd, boardDir)
}

// Get implements [board.Repository].
func (b *BoardFSRepository) Get(ctx context.Context, id string) (*board.Board, error) {
	boardDir, err := b.findBoardDir(id)
	if err != nil {
		return nil, err
	}
	return b.readBoardFromDisk(id, boardDir)
}

// Update implements [board.Repository].
// It acquires an exclusive flock on a lock file for the duration of the write,
// blocking any other process that tries to update the same board concurrently.
// The lock is released if ctx is cancelled.
func (b *BoardFSRepository) Update(ctx context.Context, id string, brd *board.Board) error {
	lockPath := filepath.Join(b.basePath, id+".lock")
	lf, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}
	defer lf.Close()
	defer os.Remove(lockPath)

	// Flock blocks until the lock is acquired. Run it in a goroutine so we
	// can respect context cancellation.
	flockDone := make(chan error, 1)
	go func() { flockDone <- syscall.Flock(int(lf.Fd()), syscall.LOCK_EX) }()
	select {
	case <-ctx.Done():
		syscall.Flock(int(lf.Fd()), syscall.LOCK_UN) // ensure unblock of the goroutine
		return ctx.Err()
	case err := <-flockDone:
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
	}
	defer syscall.Flock(int(lf.Fd()), syscall.LOCK_UN)

	boardDir := filepath.Join(b.basePath, brd.ID()+"_"+slug(brd.Name()))
	if err := b.fs.RemoveAll(boardDir); err != nil {
		return fmt.Errorf("failed to remove board dir %s: %w", boardDir, err)
	}
	return b.writeBoardToDisk(brd, boardDir)
}

func NewBoardFSRepository(fs afero.Fs, basePath string) *BoardFSRepository {
	return &BoardFSRepository{fs: fs, basePath: basePath}
}

var _ board.Repository = (*BoardFSRepository)(nil)

// findBoardDir locates the board directory for the given id.
func (b *BoardFSRepository) findBoardDir(id string) (string, error) {
	entries, err := afero.ReadDir(b.fs, b.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read boards dir: %w", err)
	}
	prefix := id + "_"
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			return filepath.Join(b.basePath, e.Name()), nil
		}
	}
	return "", fmt.Errorf("board with id %q not found", id)
}

// readBoardFromDisk reconstructs a Board from the directory at boardDir.
func (b *BoardFSRepository) readBoardFromDisk(id, boardDir string) (*board.Board, error) {
	boardMD, err := afero.ReadFile(b.fs, filepath.Join(boardDir, boardFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read board.md: %w", err)
	}
	boardTitle, _, _ := parseMarkdown(string(boardMD))

	entries, err := afero.ReadDir(b.fs, boardDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read board dir: %w", err)
	}

	type colEntry struct {
		n   int
		dto *board.UnmarshalColumnDTO
	}
	var colEntries []colEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n, _ := numericPrefix(e.Name())
		colPath := filepath.Join(boardDir, e.Name())
		colMD, err := afero.ReadFile(b.fs, filepath.Join(colPath, columnFile))
		if err != nil {
			continue
		}
		colTitle, _, _ := parseMarkdown(string(colMD))

		cardFiles, err := afero.ReadDir(b.fs, colPath)
		if err != nil {
			return nil, err
		}
		var cards []*board.UnmarshalCardDTO
		for _, cf := range cardFiles {
			if cf.IsDir() || cf.Name() == columnFile || !strings.HasSuffix(cf.Name(), ".md") {
				continue
			}
			data, err := afero.ReadFile(b.fs, filepath.Join(colPath, cf.Name()))
			if err != nil {
				return nil, err
			}
			title, headers, body := parseMarkdown(string(data))
			cardID := headers["ID"]
			if cardID == "" {
				continue
			}
			cardDTO := &board.UnmarshalCardDTO{
				ID:          cardID,
				Title:       title,
				Description: body,
			}
			if assignee, ok := headers["Assignee"]; ok && len(assignee) > 0 {
				cardDTO.Assignee = &assignee
			}
			if modifiedAtStr, ok := headers["Modified At"]; ok {
				if t, err := time.Parse(time.RFC3339, modifiedAtStr); err == nil {
					cardDTO.ModifiedAt = t
				}
			}
			cards = append(cards, cardDTO)
		}
		colEntries = append(colEntries, colEntry{n: n, dto: &board.UnmarshalColumnDTO{Name: colTitle, Cards: cards}})
	}
	slices.SortFunc(colEntries, func(a, b colEntry) int { return a.n - b.n })

	cols := make([]*board.UnmarshalColumnDTO, len(colEntries))
	for i, c := range colEntries {
		cols[i] = c.dto
	}

	return board.UnmarshalBoard(&board.UnmarshalBoardDTO{
		ID:      id,
		Name:    boardTitle,
		Columns: cols,
	}), nil
}

// writeBoardToDisk writes the full board directory structure.
func (b *BoardFSRepository) writeBoardToDisk(brd *board.Board, boardDir string) error {
	if err := b.fs.MkdirAll(boardDir, 0755); err != nil {
		return fmt.Errorf("failed to create board dir: %w", err)
	}

	// Write board.md
	boardMD := fmt.Sprintf("# %s\n\nID: %s\n\n---\n", brd.Name(), brd.ID())
	if err := afero.WriteFile(b.fs, filepath.Join(boardDir, boardFile), []byte(boardMD), 0644); err != nil {
		return err
	}

	// Write columns
	colIdx := 1
	for col := range brd.Columns() {
		colDir := filepath.Join(boardDir, fmt.Sprintf("%d_%s", colIdx, slug(col.Name())))
		if err := b.fs.MkdirAll(colDir, 0755); err != nil {
			return err
		}
		colMD := fmt.Sprintf("# %s\n\nID: %s\n\n---\n", col.Name(), slug(col.Name()))
		if err := afero.WriteFile(b.fs, filepath.Join(colDir, "column.md"), []byte(colMD), 0644); err != nil {
			return err
		}
		cardIdx := 1
		for card := range col.Cards() {
			cardFile := filepath.Join(colDir, fmt.Sprintf("%d_%s.md", cardIdx, slug(card.Title())))
			if err := afero.WriteFile(b.fs, cardFile, []byte(marshalCard(card)), 0644); err != nil {
				return err
			}
			cardIdx++
		}
		colIdx++
	}
	return nil
}

// parseMarkdown parses the fixed markdown format used for board/column/card files.
// Returns the title (from the first `# ` line), a map of header key→value pairs,
// and the body (everything after the `---` separator).
func parseMarkdown(content string) (title string, headers map[string]string, body string) {
	headers = make(map[string]string)
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) == 0 {
		return
	}
	title = strings.TrimPrefix(lines[0], "# ")
	pastSep := false
	var bodyLines []string
	for _, line := range lines[1:] {
		if pastSep {
			bodyLines = append(bodyLines, line)
			continue
		}
		if strings.HasPrefix(line, "---") {
			pastSep = true
			continue
		}
		if k, v, ok := strings.Cut(line, ": "); ok {
			headers[k] = v
		}
	}
	body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return
}

// marshalCard serializes a Card to the markdown format.
func marshalCard(c *board.Card) string {
	var sb strings.Builder
	sb.WriteString("# " + c.Title() + "\n\n")
	if c.Assignee() != nil {
		sb.WriteString("Assignee: " + *c.Assignee() + "\n")
	}
	sb.WriteString("ID: " + c.ID() + "\n")
	sb.WriteString("Modified At: " + c.ModifiedAt().Format(time.RFC3339) + "\n")
	sb.WriteString("\n---\n\n")
	sb.WriteString(c.Description())
	sb.WriteString("\n")
	return sb.String()
}

// slug converts a string to a filesystem-safe slug.
func slug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// numericPrefix extracts the leading integer from a string like "1_todo".
func numericPrefix(s string) (int, bool) {
	prefix, _, ok := strings.Cut(s, "_")
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(prefix)
	return n, err == nil
}
