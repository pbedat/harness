package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/app/query"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newReadCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a single email",
		Example: `  email read --id msg-123
  # Output:
  # Subject: Weekly update
  # From: alice@example.com
  # To: bob@example.com
  # Sent: 2025-01-15T10:30:00Z
  # Status: unread
  #
  # Hey Bob, here's this week's update...

  # Output as JSON
  email read --id msg-123 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := application.Queries.Mail.Handle(
				context.Background(),
				query.Mail{ID: id},
			)
			if err != nil {
				return err
			}

			readStatus := "unread"
			if m.ReadAt != nil {
				t, err := time.Parse(time.RFC3339, *m.ReadAt)
				if err == nil {
					readStatus = fmt.Sprintf("read at %s", t.Format(time.RFC3339))
				} else {
					readStatus = fmt.Sprintf("read at %s", *m.ReadAt)
				}
			}

			err = application.Commands.MarkRead.Handle(cmd.Context(), command.MarkRead{
				MailID: id,
			})
			if err != nil {
				log.Warn().Err(err).Str("mail_id", id).Msg("Failed to mark email as read")
			}

			if *jsonOutput {
				return writeJSON(cmd, m)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Subject: %s\nFrom: %s\nTo: %s\nSent: %s\nStatus: %s\n",
				m.Subject,
				m.From,
				strings.Join(m.To, ", "),
				m.SentAt,
				readStatus,
			)
			for _, h := range m.Headers {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", h.Name, h.Value)
			}
			body := m.Body
			if body == "" {
				body = m.HtmlBody
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", body)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Email ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
