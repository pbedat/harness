package cli

import (
	"context"
	"fmt"

	"github.com/pbedat/harness/modules/kanban/app"
	"github.com/pbedat/harness/modules/kanban/app/command"
	"github.com/spf13/cobra"
)

func newColumnCmd(application *app.Application, boardID *string) *cobra.Command {
	columnCmd := &cobra.Command{
		Use:   "column",
		Short: "Manage columns on a board",
	}
	columnCmd.AddCommand(
		newAddColumnCmd(application, boardID),
		newRemoveColumnCmd(application, boardID),
	)
	return columnCmd
}

func newAddColumnCmd(application *app.Application, boardID *string) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a column to a board",
		Example: `  kanban --board-id my-board column add --name "Code Review"
  # Output: Column "Code Review" added.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireBoardID(boardID); err != nil {
				return err
			}
			err := application.Commands.AddColumn.Handle(
				context.Background(),
				command.AddColumn{
					BoardID:    *boardID,
					ColumnName: name,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Column %q added.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Column name (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newRemoveColumnCmd(application *app.Application, boardID *string) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a column from a board (cards are migrated to the nearest column)",
		Example: `  kanban --board-id my-board column remove --name "Code Review"
  # Output: Column "Code Review" removed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireBoardID(boardID); err != nil {
				return err
			}
			err := application.Commands.RemoveColumn.Handle(
				context.Background(),
				command.RemoveColumn{
					BoardID:    *boardID,
					ColumnName: name,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Column %q removed.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Column name to remove (required)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
