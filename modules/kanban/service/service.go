package service

import (
	"os"

	"github.com/pbedat/harness/modules/kanban/adapters/out"
	"github.com/pbedat/harness/modules/kanban/app"
	"github.com/pbedat/harness/modules/kanban/app/command"
	"github.com/pbedat/harness/modules/kanban/app/query"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func NewApplication() *app.Application {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	fs := afero.NewOsFs()
	repo := out.NewBoardFSRepository(fs, ".data")

	return &app.Application{
		Commands: app.Commands{
			CreateBoard:  command.NewCreateBoardHandler(repo, logger),
			AddCard:      command.NewAddCardHandler(repo, logger),
			MoveCard:     command.NewMoveCardHandler(repo, logger),
			ArchiveCards: command.NewArchiveCardsHandler(repo, logger),
			EditCard:     command.NewEditCardHandler(repo, logger),
			AddColumn:    command.NewAddColumnHandler(repo, logger),
			RemoveColumn: command.NewRemoveColumnHandler(repo, logger),
		},
		Queries: app.Queries{
			Cards: query.NewCardsHandler(logger, repo),
			Card:  query.NewCardHandler(logger, repo),
		},
	}
}
