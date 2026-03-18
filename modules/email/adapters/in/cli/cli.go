package cli

import (
	"encoding/json"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/app"
	"github.com/spf13/cobra"
)

func Create(application *app.Application, bus *events.Bus) *cobra.Command {
	var jsonOutput bool

	rootCmd := &cobra.Command{
		Use:          "email",
		Short:        "A CLI for managing emails",
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	rootCmd.AddCommand(
		newDraftCmd(application, &jsonOutput),
		newListCmd(application, &jsonOutput),
		newMoveCmd(application, &jsonOutput),
		newQueueCmd(application, &jsonOutput),
		newReadCmd(application, &jsonOutput),
		newSendCmd(application, bus, &jsonOutput),
		newServeCmd(application),
	)

	return rootCmd
}

func writeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
