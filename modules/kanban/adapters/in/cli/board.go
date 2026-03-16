package cli

import (
	"context"
	"fmt"

	"github.com/pbedat/harness/modules/kanban/app"
	"github.com/pbedat/harness/modules/kanban/app/command"
	"github.com/spf13/cobra"
)

func newBoardCmd(application *app.Application, _ *string) *cobra.Command {
	boardCmd := &cobra.Command{
		Use:   "board",
		Short: "Manage boards",
	}
	boardCmd.AddCommand(newCreateBoardCmd(application))
	return boardCmd
}

func newCreateBoardCmd(application *app.Application) *cobra.Command {
	var (
		id      string
		name    string
		columns []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new board",
		Example: `  # Create a board with three columns
  kanban board create --id my-board --name "My Project" --columns "To Do,In Progress,Done"
  # Output: Board "my-board" created.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := application.Commands.CreateBoard.Handle(
				context.Background(),
				command.CreateBoard{
					ID:      id,
					Name:    name,
					Columns: columns,
				},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Board %q created.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Unique board ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "Human-readable board name (required)")
	cmd.Flags().StringSliceVar(&columns, "columns", nil, "Initial column names, comma-separated (required)")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("columns")

	return cmd
}
