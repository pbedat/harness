package app

import (
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	Enqueue  command.EnqueueHandler
	MarkRead command.MarkReadHandler
	Move     command.MoveHandler
}

type Queries struct {
	Mail  query.MailHandler
	Mails query.MailsHandler
}
