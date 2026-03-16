package out

import (
	"context"
	"fmt"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/email"
)

type EmailRepository struct {
	storage *MailsystemFS
	bus     *events.Bus
}

func NewEmailRepository(storage *MailsystemFS, bus *events.Bus) *EmailRepository {
	return &EmailRepository{storage: storage, bus: bus}
}

func emailHeadersToStored(headers []email.Header) []storedHeader {
	if len(headers) == 0 {
		return nil
	}
	stored := make([]storedHeader, len(headers))
	for i, h := range headers {
		stored[i] = storedHeader{Name: h.Name, Value: h.Value}
	}
	return stored
}

func storedHeadersToEmail(headers []storedHeader) []email.Header {
	if len(headers) == 0 {
		return nil
	}
	result := make([]email.Header, len(headers))
	for i, h := range headers {
		result[i] = email.Header{Name: h.Name, Value: h.Value}
	}
	return result
}

func (r *EmailRepository) Create(ctx context.Context, e *email.Email) error {
	if err := r.storage.writeEmail(&storedEmail{
		ID:        e.ID(),
		From:      e.From(),
		To:        e.To(),
		Subject:   e.Subject(),
		Body:      e.Body(),
		Headers:   emailHeadersToStored(e.Headers()),
		CreatedAt: e.CreatedAt(),
		Mailbox:   mailboxDirs[e.Mailbox()],
	}); err != nil {
		return err
	}

	for _, event := range e.Events() {
		if err := r.bus.Publish(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (r *EmailRepository) Get(ctx context.Context, id string) (*email.Email, error) {
	stored, err := r.storage.readEmail(id)
	if err != nil {
		return nil, err
	}

	mailbox, err := domain.MailboxString(stored.Mailbox)
	if err != nil {
		return nil, fmt.Errorf("invalid mailbox %q: %w", stored.Mailbox, err)
	}

	return email.UnmarshalEmail(&email.UnmarshalEmailDTO{
		ID:        stored.ID,
		From:      stored.From,
		To:        stored.To,
		Subject:   stored.Subject,
		Body:      stored.Body,
		Headers:   storedHeadersToEmail(stored.Headers),
		CreatedAt: stored.CreatedAt,
		Mailbox:   mailbox,
		ReadAt:    stored.ReadAt,
	}), nil
}

func (r *EmailRepository) Update(ctx context.Context, id string, fn func(*email.Email) error) error {
	stored, err := r.storage.readEmail(id)
	if err != nil {
		return err
	}

	oldMailbox, err := domain.MailboxString(stored.Mailbox)
	if err != nil {
		return fmt.Errorf("invalid mailbox %q: %w", stored.Mailbox, err)
	}

	e := email.UnmarshalEmail(&email.UnmarshalEmailDTO{
		ID:        stored.ID,
		From:      stored.From,
		To:        stored.To,
		Subject:   stored.Subject,
		Body:      stored.Body,
		Headers:   storedHeadersToEmail(stored.Headers),
		CreatedAt: stored.CreatedAt,
		Mailbox:   oldMailbox,
		ReadAt:    stored.ReadAt,
	})

	if err := fn(e); err != nil {
		return err
	}

	// Update stored fields from the modified entity
	stored.ReadAt = e.ReadAt()
	stored.Mailbox = mailboxDirs[e.Mailbox()]

	// If mailbox changed, delete from old location
	if e.Mailbox() != oldMailbox {
		if err := r.storage.deleteEmail(oldMailbox, id); err != nil {
			return fmt.Errorf("deleting email from old mailbox: %w", err)
		}
	}

	return r.storage.writeEmail(stored)
}
