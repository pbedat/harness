package query

import (
	"context"

	"github.com/pbedat/harness/modules/email/domain"
)

type ReadModel interface {
	GetMail(ctx context.Context, id string) (*MailReadModel, error)
	ListMails(ctx context.Context, mailbox domain.Mailbox, filterUnread *bool, limit int) ([]MailListReadModel, error)
}
