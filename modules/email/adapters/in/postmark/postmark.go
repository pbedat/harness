package postmark

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/keighl/postmark"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/pbedat/harness/modules/email/app"
	"github.com/pbedat/harness/modules/email/app/command"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/pbedat/harness/modules/email/domain/queue"
	"github.com/samber/lo"
)

func Register(application *app.Application, e *echo.Echo, username, password string) {
	h := &handler{app: application}

	g := e.Group("/webhooks/postmark",
		middleware.BasicAuth(func(c *echo.Context, user, pass string) (bool, error) {
			userOk := subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
			passOk := subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1
			return userOk && passOk, nil
		}),
	)

	g.POST("/inbound", h.handleInbound)
}

type handler struct {
	app *app.Application
}

func (h *handler) handleInbound(c *echo.Context) error {
	var msg postmark.InboundMessage
	if err := c.Bind(&msg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	createdAt, err := msg.Time()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date")
	}

	recipients := lo.Map(msg.ToFull, func(r postmark.Recipient, _ int) string {
		return r.Email
	})

	headers := make([]queue.Header, len(msg.Headers))
	for i, h := range msg.Headers {
		headers[i] = queue.Header{Name: h.Name, Value: h.Value}
	}

	dto := &queue.EmailDTO{
		ID:        msg.MessageID,
		From:      msg.FromFull.Email,
		To:        recipients,
		Subject:   msg.Subject,
		Body:      msg.TextBody,
		Headers:   headers,
		CreatedAt: createdAt,
	}

	err = h.app.Commands.Enqueue.Handle(c.Request().Context(), command.Enqueue{
		Mailbox: domain.MailboxInbox,
		Mail:    dto,
	})
	if err != nil {

		if errors.Is(err, queue.ErrDiscarded) {
			log.Info().Err(err).Str("email_id", dto.ID).Msg("Email discarded")
			return c.NoContent(http.StatusAccepted)
		}

		return err
	}

	return c.NoContent(http.StatusOK)
}
