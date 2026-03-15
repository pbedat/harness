package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/pbedat/harness/app"
	"github.com/pbedat/harness/app/command"
	"github.com/spf13/cobra"
)

func newCardCmd(application *app.Application, boardID *string) *cobra.Command {
	cardCmd := &cobra.Command{
		Use:   "card",
		Short: "Manage cards on a board",
	}
	cardCmd.AddCommand(
		newAddCardCmd(application, boardID),
		newMoveCardCmd(application, boardID),
		newEditCardCmd(application, boardID),
		newArchiveCardsCmd(application, boardID),
	)
	return cardCmd
}

func newAddCardCmd(application *app.Application, boardID *string) *cobra.Command {
	var (
		id          string
		title       string
		column      string
		description string
		assignee    string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a card to a column",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireBoardID(boardID); err != nil {
				return err
			}
			var assigneePtr *string
			if cmd.Flags().Changed("assignee") {
				assigneePtr = &assignee
			}
			err := application.Commands.AddCard.Handle(
				context.Background(),
				command.AddCard{
					BoardID:     *boardID,
					Column:      column,
					ID:          id,
					Title:       title,
					Description: description,
					Assignee:    assigneePtr,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Card %q added to column %q.\n", id, column)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Unique card ID (required)")
	cmd.Flags().StringVar(&title, "title", "", "Card title (required)")
	cmd.Flags().StringVar(&column, "column", "", "Target column name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Card description")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assignee username")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("column")

	return cmd
}

func newMoveCardCmd(application *app.Application, boardID *string) *cobra.Command {
	var (
		cardID string
		to     string
	)

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a card to a different column",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireBoardID(boardID); err != nil {
				return err
			}
			err := application.Commands.MoveCard.Handle(
				context.Background(),
				command.MoveCard{
					BoardID: *boardID,
					CardID:  cardID,
					Column:  to,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Card %q moved to %q.\n", cardID, to)
			return nil
		},
	}

	cmd.Flags().StringVar(&cardID, "id", "", "Card ID to move (required)")
	cmd.Flags().StringVar(&to, "to", "", "Destination column name (required)")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func newEditCardCmd(application *app.Application, boardID *string) *cobra.Command {
	var (
		id       string
		title    string
		body     string
		assignee string
	)

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit an existing card",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireBoardID(boardID); err != nil {
				return err
			}
			var assigneePtr *string
			if cmd.Flags().Changed("assignee") {
				assigneePtr = &assignee
			}
			err := application.Commands.EditCard.Handle(
				context.Background(),
				command.EditCard{
					BoardID:  *boardID,
					ID:       id,
					Title:    title,
					Body:     body,
					Assignee: assigneePtr,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Card %q updated.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Card ID to edit (required)")
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&body, "body", "", "New description body")
	cmd.Flags().StringVar(&assignee, "assignee", "", "New assignee username")

	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newArchiveCardsCmd(application *app.Application, boardID *string) *cobra.Command {
	var (
		column string
		stale  time.Duration
	)

	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive stale cards from a column",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireBoardID(boardID); err != nil {
				return err
			}
			err := application.Commands.ArchiveCards.Handle(
				context.Background(),
				command.ArchiveCards{
					BoardID:       *boardID,
					Column:        column,
					StaleDuration: stale,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Stale cards archived from column %q.\n", column)
			return nil
		},
	}

	cmd.Flags().StringVar(&column, "column", "", "Column to archive stale cards from (required)")
	cmd.Flags().DurationVar(&stale, "stale", 30*24*time.Hour, "Duration after which a card is considered stale")

	_ = cmd.MarkFlagRequired("column")

	return cmd
}
