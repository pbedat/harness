package out

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pbedat/harness/modules/email/domain"
	"github.com/spf13/afero"
)

// MailsystemFS is a low-level storage engine that uses the filesystem to store emails.
//
// Filesystem layout:
//
//	inbox/
//	  queue/
//	    2025-01-01T12:00:00Z_email2.json
//	  __folder.json // metadata about the folder, e.g. allowed recipients, sender, etc.
//	  2025-01-01T13:37:00Z_email1.json
//	outbox/
//	  queue/
//	    2025-01-01T12:00:00Z_email2.json
//	  __folder.json // metadata about the folder, e.g. allowed recipients, sender, etc.
//	  2025-01-01T13:37:00Z_email1.json
//	archive/
//	  2025-01-01T12:00:00Z_email2.json
//	  2025-01-01T13:37:00Z_email1.json
type MailsystemFS struct {
	fs       afero.Fs
	basePath string
}

func NewMailsystemFS(fs afero.Fs, basePath string) *MailsystemFS {
	return &MailsystemFS{
		fs:       fs,
		basePath: basePath,
	}
}

type storedHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type storedEmail struct {
	ID        string         `json:"id"`
	From      string         `json:"from"`
	To        []string       `json:"to"`
	Subject   string         `json:"subject"`
	Body      string         `json:"body"`
	HtmlBody  string         `json:"htmlBody,omitempty"`
	Headers   []storedHeader `json:"headers,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	ReadAt    *time.Time     `json:"readAt,omitempty"`
	Mailbox   string         `json:"mailbox"`
}

type folderMeta struct {
	AllowedRecipients []string `json:"allowedRecipients,omitempty"`
	AllowedFrom       *string  `json:"allowedFrom,omitempty"`
	Limit             int      `json:"limit,omitempty"`
}

var mailboxDirs = map[domain.Mailbox]string{
	domain.MailboxInbox:   "inbox",
	domain.MailboxOutbox:  "outbox",
	domain.MailboxArchive: "archive",
	domain.MailboxSent:    "sent",
}

func (m *MailsystemFS) mailboxDir(mailbox domain.Mailbox) string {
	return filepath.Join(m.basePath, mailboxDirs[mailbox])
}

func emailFileName(createdAt time.Time, id string) string {
	return fmt.Sprintf("%s_%s.json", createdAt.UTC().Format(time.RFC3339), id)
}

func (m *MailsystemFS) writeEmail(e *storedEmail) error {
	mailbox, err := domain.MailboxString(e.Mailbox)
	if err != nil {
		return fmt.Errorf("invalid mailbox %q: %w", e.Mailbox, err)
	}

	dir := m.mailboxDir(mailbox)
	if err := m.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating mailbox dir: %w", err)
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling email: %w", err)
	}

	path := filepath.Join(dir, emailFileName(e.CreatedAt, e.ID))
	return afero.WriteFile(m.fs, path, data, 0644)
}

func (m *MailsystemFS) readEmail(id string) (*storedEmail, error) {
	for _, dir := range mailboxDirs {
		emails, err := m.listEmailsInDir(filepath.Join(m.basePath, dir))
		if err != nil {
			continue
		}
		for _, e := range emails {
			if e.ID == id {
				return e, nil
			}
		}
	}
	return nil, fmt.Errorf("email %q not found", id)
}

func (m *MailsystemFS) deleteEmail(mailbox domain.Mailbox, id string) error {
	dir := m.mailboxDir(mailbox)
	entries, err := afero.ReadDir(m.fs, dir)
	if err != nil {
		return fmt.Errorf("reading mailbox dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_"+id+".json") {
			return m.fs.Remove(filepath.Join(dir, entry.Name()))
		}
	}
	return fmt.Errorf("email %q not found in %s", id, mailboxDirs[mailbox])
}

func (m *MailsystemFS) listEmails(mailbox domain.Mailbox) ([]*storedEmail, error) {
	return m.listEmailsInDir(m.mailboxDir(mailbox))
}

func (m *MailsystemFS) listEmailsInDir(dir string) ([]*storedEmail, error) {
	entries, err := afero.ReadDir(m.fs, dir)
	if err != nil {
		if exists, _ := afero.DirExists(m.fs, dir); !exists {
			return nil, nil
		}
		return nil, fmt.Errorf("reading dir: %w", err)
	}

	var emails []*storedEmail
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || entry.Name() == "__folder.json" {
			continue
		}

		data, err := afero.ReadFile(m.fs, filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading email file %s: %w", entry.Name(), err)
		}

		var e storedEmail
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, fmt.Errorf("parsing email file %s: %w", entry.Name(), err)
		}
		emails = append(emails, &e)
	}
	return emails, nil
}

func (m *MailsystemFS) readQueueMeta(mailbox domain.Mailbox) (*folderMeta, error) {
	path := filepath.Join(m.mailboxDir(mailbox), "__folder.json")
	data, err := afero.ReadFile(m.fs, path)
	if err != nil {
		if exists, _ := afero.Exists(m.fs, path); !exists {
			return &folderMeta{}, nil
		}
		return nil, fmt.Errorf("reading folder meta: %w", err)
	}

	var meta folderMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing folder meta: %w", err)
	}
	return &meta, nil
}

func (m *MailsystemFS) readQueueEmails(mailbox domain.Mailbox) ([]*storedEmail, error) {
	queueDir := filepath.Join(m.mailboxDir(mailbox), "queue")
	return m.listEmailsInDir(queueDir)
}

func (m *MailsystemFS) writeQueueEmail(mailbox domain.Mailbox, e *storedEmail) error {
	queueDir := filepath.Join(m.mailboxDir(mailbox), "queue")

	if err := m.fs.MkdirAll(queueDir, 0755); err != nil {
		return fmt.Errorf("creating queue dir: %w", err)
	}

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling queue email: %w", err)
	}

	path := filepath.Join(queueDir, emailFileName(e.CreatedAt, e.ID))
	return afero.WriteFile(m.fs, path, data, 0644)
}

func (m *MailsystemFS) deleteQueueEmail(mailbox domain.Mailbox, id string) error {
	queueDir := filepath.Join(m.mailboxDir(mailbox), "queue")
	entries, err := afero.ReadDir(m.fs, queueDir)
	if err != nil {
		return fmt.Errorf("reading queue dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_"+id+".json") {
			return m.fs.Remove(filepath.Join(queueDir, entry.Name()))
		}
	}
	return fmt.Errorf("queue email %q not found in %s", id, mailboxDirs[mailbox])
}
