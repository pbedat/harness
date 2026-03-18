package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/query"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/spf13/cobra"
)

func newListCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var (
		mailbox      string
		filterUnread *bool
		limit        int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List emails in a mailbox",
		Example: `  email list --mailbox inbox
  # Output:
  # [msg-123] Weekly update  From: alice@example.com  To: bob@example.com
  # [msg-124] Bug report     From: carol@example.com  To: bob@example.com

  # Filter to unread only
  email list --mailbox inbox --unread

  # Limit results and output as JSON
  email list --mailbox inbox --limit 10 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mb, err := domain.MailboxString(mailbox)
			if err != nil {
				return fmt.Errorf("invalid mailbox %q: %w", mailbox, err)
			}

			if cmd.Flags().Changed("unread") {
				val, _ := cmd.Flags().GetBool("unread")
				filterUnread = &val
			}

			mails, err := application.Queries.Mails.Handle(
				context.Background(),
				query.Mails{
					Mailbox:      mb,
					FilterUnread: filterUnread,
					Limit:        limit,
				},
			)
			if err != nil {
				return err
			}

			if *jsonOutput {
				return writeJSON(cmd, mails)
			}

			for _, m := range mails {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s  From: %s  To: %s\n",
					m.ID,
					m.Subject,
					m.From,
					strings.Join(m.To, ", "),
				)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mailbox, "mailbox", "", "Mailbox to list (required)")
	cmd.Flags().Bool("unread", false, "Filter to unread emails only")
	cmd.Flags().IntVar(&limit, "limit", 0, "Max emails to return (0 = all)")
	_ = cmd.MarkFlagRequired("mailbox")

	return cmd
}
