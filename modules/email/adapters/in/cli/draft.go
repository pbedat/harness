package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/spf13/cobra"
)

func newDraftCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var (
		id         string
		from       string
		to         []string
		subject    string
		body       string
		bodyFile   string
		replyTo    string
		replyingTo string
	)

	cmd := &cobra.Command{
		Use:   "draft",
		Short: "Create an email draft",
		Example: `  email draft --from me@example.com --to alice@example.com --subject "Hello" --body "Hi there"
  email draft --from me@example.com --to alice@example.com --subject "Hello" --body-file ./message.txt
  email draft --from me@example.com --to alice@example.com --subject "Re: Bug" --body "Fixed" --replying-to msg-123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("body-file") {
				data, err := os.ReadFile(bodyFile)
				if err != nil {
					return fmt.Errorf("reading body file: %w", err)
				}
				body = string(data)
			}

			draft := command.Draft{
				EmailID: id,
				From:    from,
				To:      to,
				Subject: subject,
				Body:    body,
			}

			if replyTo != "" {
				draft.ReplyTo = &replyTo
			}
			if replyingTo != "" {
				draft.ReplyingToEmailID = &replyingTo
			}

			err := application.Commands.Draft.Handle(context.Background(), draft)
			if err != nil {
				return err
			}

			if *jsonOutput {
				return writeJSON(cmd, map[string]string{"status": "drafted"})
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Draft created.")
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "email id (required)")
	cmd.Flags().StringVar(&from, "from", "", "Sender email address (required)")
	cmd.Flags().StringSliceVar(&to, "to", nil, "Recipient email addresses (required, comma-separated)")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&body, "body", "", "Email body text")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Read email body from file")
	cmd.Flags().StringVar(&replyTo, "reply-to", "", "Reply-To address")
	cmd.Flags().StringVar(&replyingTo, "replying-to", "", "Email ID being replied to")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("id")
	cmd.MarkFlagsMutuallyExclusive("body", "body-file")
	cmd.MarkFlagsOneRequired("body", "body-file")

	return cmd
}
