package query

import (
	"context"

	"github.com/pbedat/harness/common/decorator"
	"github.com/rs/zerolog"
)

type Mail struct {
	ID string
}

type Header struct {
	Name  string
	Value string
}

type MailReadModel struct {
	Subject  string
	From     string
	To       []string
	Body     string
	HtmlBody string
	Headers  []Header
	SentAt   string
	ReadAt   *string
}

type MailHandler decorator.QueryHandler[Mail, *MailReadModel]

func NewMailHandler(
	logger zerolog.Logger,
	readModel ReadModel,
) MailHandler {
	return decorator.ApplyQueryDecorators(
		&mailHandler{readModel: readModel}, logger,
	)
}

type mailHandler struct {
	readModel ReadModel
}

func (h mailHandler) Handle(ctx context.Context, q Mail) (*MailReadModel, error) {
	return h.readModel.GetMail(ctx, q.ID)
}
