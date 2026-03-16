package postmark_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/pbedat/harness/modules/email/adapters/in/postmark"
	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/domain"
)

type spyEnqueueHandler struct {
	called bool
	cmd    command.Enqueue
	err    error
}

func (s *spyEnqueueHandler) Handle(_ context.Context, cmd command.Enqueue) error {
	s.called = true
	s.cmd = cmd
	return s.err
}

const validPayload = `{
	"FromFull": {"Email": "sender@example.com", "Name": "Sender"},
	"ToFull": [{"Email": "recipient@example.com", "Name": "Recipient"}],
	"Subject": "Test Subject",
	"TextBody": "Hello, this is a test email.",
	"Headers": [{"Name": "X-Spam-Status", "Value": "No"}, {"Name": "Received", "Value": "from mx1.example.com"}],
	"MessageID": "msg-123",
	"Date": "Thu, 05 Jan 2023 15:04:05 +0000"
}`

func setupTest(enqueueHandler command.EnqueueHandler) *echo.Echo {
	application := &app.Application{
		Commands: app.Commands{
			Enqueue: enqueueHandler,
		},
	}

	e := echo.New()
	postmark.Register(application, e, "webhook-user", "webhook-pass")
	return e
}

func basicAuth(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func TestHandleInbound(t *testing.T) {
	t.Run("enqueues email to inbox", func(t *testing.T) {
		spy := &spyEnqueueHandler{}
		e := setupTest(spy)

		req := httptest.NewRequest(http.MethodPost, "/webhooks/postmark/inbound", strings.NewReader(validPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, basicAuth("webhook-user", "webhook-pass"))
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
		}

		if !spy.called {
			t.Fatal("expected enqueue handler to be called")
		}

		if spy.cmd.Mailbox != domain.MailboxInbox {
			t.Errorf("expected mailbox Inbox, got %v", spy.cmd.Mailbox)
		}
		if spy.cmd.Mail.ID != "msg-123" {
			t.Errorf("expected ID 'msg-123', got %q", spy.cmd.Mail.ID)
		}
		if spy.cmd.Mail.From != "sender@example.com" {
			t.Errorf("expected from 'sender@example.com', got %q", spy.cmd.Mail.From)
		}
		if len(spy.cmd.Mail.To) != 1 || spy.cmd.Mail.To[0] != "recipient@example.com" {
			t.Errorf("expected to ['recipient@example.com'], got %v", spy.cmd.Mail.To)
		}
		if spy.cmd.Mail.Subject != "Test Subject" {
			t.Errorf("expected subject 'Test Subject', got %q", spy.cmd.Mail.Subject)
		}
		if spy.cmd.Mail.Body != "Hello, this is a test email." {
			t.Errorf("expected body 'Hello, this is a test email.', got %q", spy.cmd.Mail.Body)
		}
		if len(spy.cmd.Mail.Headers) != 2 {
			t.Fatalf("expected 2 headers, got %d", len(spy.cmd.Mail.Headers))
		}
		if spy.cmd.Mail.Headers[0].Name != "X-Spam-Status" || spy.cmd.Mail.Headers[0].Value != "No" {
			t.Errorf("expected first header X-Spam-Status: No, got %s: %s", spy.cmd.Mail.Headers[0].Name, spy.cmd.Mail.Headers[0].Value)
		}
		if spy.cmd.Mail.Headers[1].Name != "Received" || spy.cmd.Mail.Headers[1].Value != "from mx1.example.com" {
			t.Errorf("expected second header Received: from mx1.example.com, got %s: %s", spy.cmd.Mail.Headers[1].Name, spy.cmd.Mail.Headers[1].Value)
		}
	})

	t.Run("returns 401 without credentials", func(t *testing.T) {
		spy := &spyEnqueueHandler{}
		e := setupTest(spy)

		req := httptest.NewRequest(http.MethodPost, "/webhooks/postmark/inbound", strings.NewReader(validPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
		if spy.called {
			t.Fatal("enqueue handler should not have been called")
		}
	})

	t.Run("returns 401 with wrong credentials", func(t *testing.T) {
		spy := &spyEnqueueHandler{}
		e := setupTest(spy)

		req := httptest.NewRequest(http.MethodPost, "/webhooks/postmark/inbound", strings.NewReader(validPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, basicAuth("wrong", "creds"))
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
		if spy.called {
			t.Fatal("enqueue handler should not have been called")
		}
	})

	t.Run("returns 400 for invalid payload", func(t *testing.T) {
		spy := &spyEnqueueHandler{}
		e := setupTest(spy)

		req := httptest.NewRequest(http.MethodPost, "/webhooks/postmark/inbound", strings.NewReader("not json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, basicAuth("webhook-user", "webhook-pass"))
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})
}
