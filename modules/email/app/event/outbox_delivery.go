package event

import (
	"context"
	"time"

	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/email"
	"github.com/pbedat/harness/modules/email/domain/queue"
)

type OutboxDeliveryHandler struct {
	queueRepo queue.Repository
	repo      email.Repository
	mailPort  mailPort
}

type mailPort interface {
	Send(ctx context.Context, mail *queue.Email) error
}

func NewOutboxDeliveryHandler(queueRepo queue.Repository, repo email.Repository, mailPort mailPort) *OutboxDeliveryHandler {
	return &OutboxDeliveryHandler{queueRepo: queueRepo, repo: repo, mailPort: mailPort}
}

func (h *OutboxDeliveryHandler) Handle(ctx context.Context, event any) error {
	switch ev := event.(type) {

	case *email.MovedEvent:
		if ev.Mailbox != domain.MailboxOutbox {
			return nil
		}
		return h.queueRepo.Update(ctx, domain.MailboxSent, func(q *queue.Queue) error {
			dto := queue.EmailDTO{
				ID:        ev.EmailID,
				From:      ev.From,
				To:        ev.To,
				Subject:   ev.Subject,
				Body:      ev.Body,
				HtmlBody:  ev.HtmlBody,
				Headers:   make([]queue.Header, len(ev.Headers)),
				CreatedAt: time.Now(),
			}

			m, err := dto.ToEmail()
			if err != nil {
				return err
			}

			return q.Enqueue(m)
		})
	case *queue.EnqueuedEvent:
		return h.handleOutboxDelivery(ctx, ev)
	default:
		return nil
	}

}

func (h *OutboxDeliveryHandler) handleOutboxDelivery(ctx context.Context, ev *queue.EnqueuedEvent) error {
	return h.queueRepo.Update(ctx, domain.MailboxSent, func(q *queue.Queue) error {
		queueMail, ok := q.Dequeue()
		if !ok {
			return nil
		}

		if err := h.mailPort.Send(ctx, queueMail); err != nil {
			return err
		}

		queueHeaders := queueMail.Headers()
		emailHeaders := make([]email.Header, len(queueHeaders))
		for i, h := range queueHeaders {
			emailHeaders[i] = email.Header{Name: h.Name, Value: h.Value}
		}

		mail, err := email.Erstellen(&email.NewEmailDTO{
			ID:        queueMail.ID(),
			Mailbox:   domain.MailboxSent,
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

		if err := h.repo.Update(ctx, mail.ID(), func(e *email.Email) error {
			return e.Move(domain.MailboxSent)
		}); err != nil {
			return err
		}

		return nil
	})
}
