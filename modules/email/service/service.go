package service

import (
	"os"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/adapters/out"
	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	appevent "github.com/pbedat/harness/modules/email/app/event"
	"github.com/pbedat/harness/modules/email/app/query"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func NewApplication(bus *events.Bus, fs afero.Fs, basePath string, postmarkServerToken func() string) *app.Application {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	storage := out.NewMailsystemFS(fs, basePath)

	emailRepo := out.NewEmailRepository(storage, bus)
	queueRepo := out.NewQueueRepository(storage, bus)
	readModel := out.NewReadModelAdapter(storage)

	inboxDelivery := appevent.NewInboxDeliveryHandler(emailRepo, queueRepo)
	bus.Subscribe(inboxDelivery.Handle)

	mailAdapter := out.NewPostmarkMailAdapter(postmarkServerToken)
	outboxDelivery := appevent.NewOutboxDeliveryHandler(queueRepo, emailRepo, mailAdapter)
	bus.Subscribe(outboxDelivery.Handle)

	return &app.Application{
		Commands: app.Commands{
			Draft:          command.NewDraftHandler(emailRepo, logger),
			Enqueue:        command.NewEnqueueHandler(queueRepo, logger),
			MarkRead:       command.NewMarkReadHandler(emailRepo, logger),
			Move:           command.NewMoveHandler(emailRepo, logger),
			ConfigureQueue: command.NewConfigureQueueHandler(storage, logger),
		},
		Queries: app.Queries{
			Mail:  query.NewMailHandler(logger, readModel),
			Mails: query.NewMailsHandler(logger, readModel),
		},
	}
}
