package cli

import (
	"context"
	"fmt"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/spf13/cobra"
)

func newMoveCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var (
		id string
		to string
	)

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move an email to a different mailbox",
		Example: `  email move --id msg-123 --to archive
  # Output: Email "msg-123" moved to "archive".

  # Output as JSON
  email move --id msg-123 --to trash --json
  # Output: {"id": "msg-123", "movedTo": "trash"}`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mb, err := domain.MailboxString(to)
			if err != nil {
				return fmt.Errorf("invalid mailbox %q: %w", to, err)
			}

			err = application.Commands.Move.Handle(
				context.Background(),
				command.Move{
					MailID: id,
					To:     mb,
				},
			)
			if err != nil {
				return err
			}

			if *jsonOutput {
				return writeJSON(cmd, map[string]string{
					"id":      id,
					"movedTo": to,
				})
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Email %q moved to %q.\n", id, to)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Email ID to move (required)")
	cmd.Flags().StringVar(&to, "to", "", "Destination mailbox (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}
