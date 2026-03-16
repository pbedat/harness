package out

import (
	"context"
	"sort"

	"github.com/pbedat/harness/modules/email/app/query"
	"github.com/pbedat/harness/modules/email/domain"
)

type ReadModelAdapter struct {
	storage *MailsystemFS
}

func NewReadModelAdapter(storage *MailsystemFS) *ReadModelAdapter {
	return &ReadModelAdapter{storage: storage}
}

func (r *ReadModelAdapter) GetMail(ctx context.Context, id string) (*query.MailReadModel, error) {
	stored, err := r.storage.readEmail(id)
	if err != nil {
		return nil, err
	}

	var readAt *string
	if stored.ReadAt != nil {
		s := stored.ReadAt.String()
		readAt = &s
	}

	var headers []query.Header
	for _, h := range stored.Headers {
		headers = append(headers, query.Header{Name: h.Name, Value: h.Value})
	}

	return &query.MailReadModel{
		Subject: stored.Subject,
		From:    stored.From,
		To:      stored.To,
		Body:    stored.Body,
		Headers: headers,
		SentAt:  stored.CreatedAt.String(),
		ReadAt:  readAt,
	}, nil
}

func (r *ReadModelAdapter) ListMails(ctx context.Context, mailbox domain.Mailbox, filterUnread *bool, limit int) ([]query.MailListReadModel, error) {
	emails, err := r.storage.listEmails(mailbox)
	if err != nil {
		return nil, err
	}

	// Sort by creation time descending (newest first)
	sort.Slice(emails, func(i, j int) bool {
		return emails[i].CreatedAt.After(emails[j].CreatedAt)
	})

	var result []query.MailListReadModel
	for _, e := range emails {
		if filterUnread != nil {
			isUnread := e.ReadAt == nil
			if *filterUnread != isUnread {
				continue
			}
		}

		result = append(result, query.MailListReadModel{
			ID:         e.ID,
			ReceivedAt: e.CreatedAt,
			From:       e.From,
			Subject:    e.Subject,
			To:         e.To,
		})

		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result, nil
}
