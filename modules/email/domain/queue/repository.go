package queue

import (
	"context"

	"github.com/pbedat/harness/modules/email/domain"
)

type Repository interface {
	Update(ctx context.Context, mailbox domain.Mailbox, fn func(*Queue) error) error
}
