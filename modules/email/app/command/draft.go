package command

import (
	"context"
	"time"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/email"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type Draft struct {
	EmailID string
	Subject string
	Body    string
	To      []string
	From    string
	ReplyTo *string

	ReplyingToEmailID *string
}

type DraftHandler decorator.CommandHandler[Draft]

func NewDraftHandler(
	emailRepo email.Repository,
	logger zerolog.Logger,
) DraftHandler {
	return decorator.ApplyCommandDecorators(
		&draftHandler{emailRepo: emailRepo}, logger,
	)
}

type draftHandler struct {
	emailRepo email.Repository
}

func (h draftHandler) Handle(ctx context.Context, cmd Draft) error {

	draft := email.NewEmailDTO{
		ID:        cmd.EmailID,
		Mailbox:   domain.MailboxDraft,
		From:      cmd.From,
		To:        cmd.To,
		Subject:   cmd.Subject,
		Body:      cmd.Body,
		CreatedAt: time.Now(),
	}

	if cmd.ReplyingToEmailID != nil {
		m, err := h.emailRepo.Get(ctx, *cmd.ReplyingToEmailID)
		if err != nil {
			return err
		}

		messageID, ok := lo.Find(m.Headers(), func(h email.Header) bool {
			return h.Name == "Message-ID"
		})

		if ok {
			draft.Headers = append(draft.Headers, email.Header{
				Name:  "In-Reply-To",
				Value: messageID.Value,
			})
		}
	}

	m, err := draft.ToEmail()
	if err != nil {
		return err
	}

	if err := h.emailRepo.Create(ctx, m); err != nil {
		return err
	}

	return nil
}
