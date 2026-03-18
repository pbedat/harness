package cli

import (
	"context"
	"fmt"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/spf13/cobra"
)

func newQueueCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	queueCmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage email queues",
	}

	queueCmd.AddCommand(
		newQueueConfigCmd(application, jsonOutput),
	)

	return queueCmd
}

func newQueueConfigCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var (
		mailbox           string
		allowedRecipients []string
		allowedFrom       string
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure queue allow-lists for a mailbox",
		Example: `  email queue config --mailbox Inbox --allowed-recipients "pat@copilot.events" --allowed-from "support@copilot.events"
  email queue config --mailbox Outbox --allowed-recipients "alice@example.com,bob@example.com"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mb, err := domain.MailboxString(mailbox)
			if err != nil {
				return fmt.Errorf("invalid mailbox %q: %w", mailbox, err)
			}

			configure := command.ConfigureQueue{
				Mailbox:           mb,
				AllowedRecipients: allowedRecipients,
			}

			if cmd.Flags().Changed("allowed-from") {
				configure.AllowedFrom = &allowedFrom
			}

			if err := application.Commands.ConfigureQueue.Handle(context.Background(), configure); err != nil {
				return err
			}

			if *jsonOutput {
				return writeJSON(cmd, map[string]string{"status": "configured"})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Queue configured.")
			return nil
		},
	}

	cmd.Flags().StringVar(&mailbox, "mailbox", "", "Mailbox to configure (Inbox or Outbox)")
	cmd.Flags().StringSliceVar(&allowedRecipients, "allowed-recipients", nil, "Allowed recipient email addresses (comma-separated)")
	cmd.Flags().StringVar(&allowedFrom, "allowed-from", "", "Allowed sender email address")
	_ = cmd.MarkFlagRequired("mailbox")

	return cmd
}
