package app

import (
	"github.com/pbedat/harness/modules/kanban/app/command"
	"github.com/pbedat/harness/modules/kanban/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateBoard  command.CreateBoardHandler
	AddCard      command.AddCardHandler
	MoveCard     command.MoveCardHandler
	ArchiveCards command.ArchiveCardsHandler
	EditCard     command.EditCardHandler
	AddColumn    command.AddColumnHandler
	RemoveColumn command.RemoveColumnHandler
}

type Queries struct {
	Cards query.CardsHandler
	Card  query.CardHandler
}
