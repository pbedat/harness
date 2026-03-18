package out

import (
	"context"
	"strings"

	"github.com/keighl/postmark"
	"github.com/pbedat/harness/modules/email/domain/queue"
)

type PostmarkMailAdapter struct {
	serverToken func() string
}

func NewPostmarkMailAdapter(serverToken func() string) *PostmarkMailAdapter {
	return &PostmarkMailAdapter{serverToken: serverToken}
}

func (a *PostmarkMailAdapter) Send(_ context.Context, mail *queue.Email) error {
	client := postmark.NewClient(a.serverToken(), "")

	queueHeaders := mail.Headers()
	headers := make([]postmark.Header, len(queueHeaders))
	for i, h := range queueHeaders {
		headers[i] = postmark.Header{Name: h.Name, Value: h.Value}
	}

	_, err := client.SendEmail(postmark.Email{
		From:     mail.From(),
		To:       strings.Join(mail.To(), ","),
		Subject:  mail.Subject(),
		TextBody: mail.Body(),
		HtmlBody: mail.HtmlBody(),
		Headers:  headers,
	})
	return err
}
