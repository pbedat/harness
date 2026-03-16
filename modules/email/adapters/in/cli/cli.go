package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/app/query"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func Create(application *app.Application) *cobra.Command {
	var jsonOutput bool

	rootCmd := &cobra.Command{
		Use:          "email",
		Short:        "A CLI for managing emails",
		SilenceUsage: true,
	}

	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	rootCmd.AddCommand(
		newListCmd(application, &jsonOutput),
		newReadCmd(application, &jsonOutput),
		newMoveCmd(application, &jsonOutput),
		newServeCmd(application),
	)

	return rootCmd
}

func writeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newListCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var (
		mailbox      string
		filterUnread *bool
		limit        int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List emails in a mailbox",
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

func newReadCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a single email",
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
			fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", m.Body)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Email ID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

func newMoveCmd(application *app.Application, jsonOutput *bool) *cobra.Command {
	var (
		id string
		to string
	)

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move an email to a different mailbox",
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
					"id":     id,
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
