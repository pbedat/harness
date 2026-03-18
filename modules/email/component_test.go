package email_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/spf13/afero"

	events "github.com/pbedat/harness/common/event"
	"github.com/pbedat/harness/modules/email/adapters/in/postmark"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/email"
	"github.com/pbedat/harness/modules/email/service"
)

func TestPostmarkToInbox(t *testing.T) {
	bus := events.NewBus()

	erstellt := make(chan email.ErstelltEvent, 1)
	bus.Subscribe(func(_ context.Context, event any) error {
		if e, ok := event.(email.ErstelltEvent); ok {
			erstellt <- e
		}
		return nil
	})

	application := service.NewApplication(bus, afero.NewMemMapFs(), ".maildata", func() string { return "" })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus.Run(ctx)

	e := echo.New()
	postmark.Register(application, e, "webhook-user", "webhook-pass")

	t.Run("postmark inbound email is delivered to inbox", func(t *testing.T) {
		payload := `{
			"FromFull": {"Email": "sender@example.com", "Name": "Sender"},
			"ToFull": [{"Email": "recipient@example.com", "Name": "Recipient"}],
			"Subject": "Test Subject",
			"TextBody": "Hello, this is a test email.",
			"Headers": [{"Name": "X-Spam-Status", "Value": "No"}],
			"MessageID": "msg-123",
			"Date": "Thu, 05 Jan 2023 15:04:05 +0000"
		}`

		req := httptest.NewRequest(http.MethodPost, "/webhooks/postmark/inbound", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte("webhook-user:webhook-pass")))
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		select {
		case event := <-erstellt:
			if event.Email.ID != "msg-123" {
				t.Errorf("expected ID %q, got %q", "msg-123", event.Email.ID)
			}
			if event.Email.From != "sender@example.com" {
				t.Errorf("expected from %q, got %q", "sender@example.com", event.Email.From)
			}
			if event.Email.Subject != "Test Subject" {
				t.Errorf("expected subject %q, got %q", "Test Subject", event.Email.Subject)
			}
			if len(event.Email.To) != 1 || event.Email.To[0] != "recipient@example.com" {
				t.Errorf("expected to [recipient@example.com], got %v", event.Email.To)
			}
			if event.Email.Mailbox != domain.MailboxInbox {
				t.Errorf("expected mailbox Inbox, got %v", event.Email.Mailbox)
			}
			if len(event.Email.Headers) != 1 {
				t.Fatalf("expected 1 header, got %d", len(event.Email.Headers))
			}
			if event.Email.Headers[0].Name != "X-Spam-Status" || event.Email.Headers[0].Value != "No" {
				t.Errorf("expected header X-Spam-Status: No, got %s: %s", event.Email.Headers[0].Name, event.Email.Headers[0].Value)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for ErstelltEvent")
		}
	})
}
