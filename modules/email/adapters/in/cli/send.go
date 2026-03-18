package cli

import (
	"context"
	"fmt"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/spf13/cobra"
)

func newSendCmd(application *app.Application, bus *events.Bus, jsonOutput *bool) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an existing draft",
		Example: `  email send --id draft-123
  # Output: Email "draft-123" sent.

  email send --id draft-123 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := application.Commands.Move.Handle(
				context.Background(),
				command.Move{
					MailID: id,
					To:     domain.MailboxOutbox,
				},
			)
			if err != nil {
				return err
			}

			bus.Drain()

			if *jsonOutput {
				return writeJSON(cmd, map[string]string{
					"id":     id,
					"status": "sent",
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Email %q sent.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Draft email ID to send (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
