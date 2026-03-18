package event

import (
	"context"

	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/email"
	"github.com/pbedat/harness/modules/email/domain/queue"
)

type InboxDeliveryHandler struct {
	repo      email.Repository
	queueRepo queue.Repository
}

func NewInboxDeliveryHandler(repo email.Repository, queueRepo queue.Repository) *InboxDeliveryHandler {
	return &InboxDeliveryHandler{repo: repo, queueRepo: queueRepo}
}

func (h *InboxDeliveryHandler) Handle(ctx context.Context, event any) error {
	_, ok := event.(*queue.EnqueuedEvent)
	if !ok {
		return nil
	}

	return h.queueRepo.Update(ctx, domain.MailboxInbox, func(q *queue.Queue) error {
		queueMail, ok := q.Dequeue()
		if !ok {
			return nil
		}

		queueHeaders := queueMail.Headers()
		emailHeaders := make([]email.Header, len(queueHeaders))
		for i, h := range queueHeaders {
			emailHeaders[i] = email.Header{Name: h.Name, Value: h.Value}
		}

		mail, err := email.Erstellen(&email.NewEmailDTO{
			ID:        queueMail.ID(),
			Mailbox:   domain.MailboxInbox,
			From:      queueMail.From(),
			To:        queueMail.To(),
			Subject:   queueMail.Subject(),
			Body:      queueMail.Body(),
			HtmlBody:  queueMail.HtmlBody(),
			Headers:   emailHeaders,
			CreatedAt: queueMail.CreatedAt(),
		})
		if err != nil {
			return err
		}

		if err := h.repo.Create(ctx, mail); err != nil {
			return err
		}

		return nil
	})
}
