package out

import (
	"context"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/queue"
)

func queueHeadersToStored(headers []queue.Header) []storedHeader {
	if len(headers) == 0 {
		return nil
	}
	stored := make([]storedHeader, len(headers))
	for i, h := range headers {
		stored[i] = storedHeader{Name: h.Name, Value: h.Value}
	}
	return stored
}

func storedHeadersToQueue(headers []storedHeader) []queue.Header {
	if len(headers) == 0 {
		return nil
	}
	result := make([]queue.Header, len(headers))
	for i, h := range headers {
		result[i] = queue.Header{Name: h.Name, Value: h.Value}
	}
	return result
}

type QueueRepository struct {
	storage *MailsystemFS
	bus     *events.Bus
}

func NewQueueRepository(storage *MailsystemFS, bus *events.Bus) *QueueRepository {
	return &QueueRepository{storage: storage, bus: bus}
}

func (r *QueueRepository) Update(ctx context.Context, mailbox domain.Mailbox, fn func(*queue.Queue) error) error {
	meta, err := r.storage.readQueueMeta(mailbox)
	if err != nil {
		return err
	}

	queueEmails, err := r.storage.readQueueEmails(mailbox)
	if err != nil {
		return err
	}

	// Convert stored emails to domain emails
	domainEmails := make([]*queue.Email, 0, len(queueEmails))
	for _, se := range queueEmails {
		dto := &queue.EmailDTO{
			ID:        se.ID,
			From:      se.From,
			To:        se.To,
			Subject:   se.Subject,
			Body:      se.Body,
			HtmlBody:  se.HtmlBody,
			Headers:   storedHeadersToQueue(se.Headers),
			CreatedAt: se.CreatedAt,
		}
		e, err := dto.ToEmail()
		if err != nil {
			return err
		}
		domainEmails = append(domainEmails, e)
	}

	q := queue.UnmarshalQueue(&queue.UnmarshalQueueDTO{
		Mailbox:           mailbox,
		AllowedRecipients: meta.AllowedRecipients,
		AllowedFrom:       meta.AllowedFrom,
		Limit:             meta.Limit,
		Emails:            domainEmails,
	})

	if err := fn(q); err != nil {
		return err
	}

	// Apply filesystem changes based on domain events
	for _, event := range q.Events() {
		switch e := event.(type) {
		case *queue.EnqueuedEvent:
			if err := r.storage.writeQueueEmail(mailbox, &storedEmail{
				ID:        e.Email.ID,
				From:      e.Email.From,
				To:        e.Email.To,
				Subject:   e.Email.Subject,
				Body:      e.Email.Body,
				HtmlBody:  e.Email.HtmlBody,
				Headers:   queueHeadersToStored(e.Email.Headers),
				CreatedAt: e.Email.CreatedAt,
				Mailbox:   mailboxDirs[mailbox],
			}); err != nil {
				return err
			}
		case *queue.DequeuedEvent:
			if err := r.storage.deleteQueueEmail(mailbox, e.EmailID); err != nil {
				return err
			}
		}

		if err := r.bus.Publish(ctx, event); err != nil {
			return err
		}
	}

	return nil
}
