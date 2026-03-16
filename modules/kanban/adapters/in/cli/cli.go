package cli

import (
	"fmt"

	"github.com/pbedat/harness/modules/kanban/app"
	"github.com/spf13/cobra"
)

func Create(application *app.Application) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "kanban",
		Short:        "A CLI for managing Kanban boards",
		SilenceUsage: true,
	}

	var boardID string
	rootCmd.PersistentFlags().StringVar(&boardID, "board-id", "", "ID of the board to operate on")

	rootCmd.AddCommand(
		newBoardCmd(application, &boardID),
		newCardCmd(application, &boardID),
		newColumnCmd(application, &boardID),
	)

	return rootCmd
}

func requireBoardID(boardID *string) error {
	if *boardID == "" {
		return fmt.Errorf("--board-id is required: specify the board to operate on")
	}
	return nil
}
