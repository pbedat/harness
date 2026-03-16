package query

import (
	"context"
	"time"

	"github.com/pbedat/harness/common/decorator"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/rs/zerolog"
)

type Mails struct {
	Mailbox      domain.Mailbox
	FilterUnread *bool
	Limit        int
}

type MailListReadModel struct {
	ID         string
	ReceivedAt time.Time
	From       string
	Subject    string
	To         []string
}

type MailsHandler decorator.QueryHandler[Mails, []MailListReadModel]

func NewMailsHandler(
	logger zerolog.Logger,
	readModel ReadModel,
) MailsHandler {
	return decorator.ApplyQueryDecorators(
		&mailsHandler{readModel: readModel}, logger,
	)
}

type mailsHandler struct {
	readModel ReadModel
}

func (h mailsHandler) Handle(ctx context.Context, q Mails) ([]MailListReadModel, error) {
	return h.readModel.ListMails(ctx, q.Mailbox, q.FilterUnread, q.Limit)
}
