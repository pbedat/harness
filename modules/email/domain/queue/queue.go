package queue

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pbedat/harness/modules/email/domain"
	"github.com/samber/lo"
)

func NewQueue(mailbox domain.Mailbox) *Queue {

	switch mailbox {
	case domain.MailboxInbox, domain.MailboxOutbox:
		// ok
	default:
		panic("invalid mailbox: " + mailbox.String())
	}

	return &Queue{
		mailbox: mailbox,
	}
}

type Queue struct {
	mailbox           domain.Mailbox
	allowedRecipients []string
	allowedFrom       *string
	limit             int

	emails []*Email

	events []any
}

var ErrDiscarded = errors.New("email discarded")
var ErrFull = errors.New("queue full")

func (q *Queue) Enqueue(email *Email) error {
	if q.allowedFrom != nil && email.from != *q.allowedFrom {
		// Sender not allowed, reject email
		return fmt.Errorf("%w: sender not allowed: %s", ErrDiscarded, email.from)
	}

	if len(q.allowedRecipients) > 0 {
		if !lo.Every(q.allowedRecipients, email.to) {
			return fmt.Errorf("%w: recipients not allowed: %v", ErrDiscarded, lo.Without(email.to, q.allowedRecipients...))
		}
	}

	if q.limit > 0 && len(q.emails) >= q.limit {
		return fmt.Errorf("%w: queue limit reached: %d", ErrFull, q.limit)
	}

	q.emails = append(q.emails, email)
	q.events = append(q.events, &EnqueuedEvent{
		Email: &EmailDTO{
			ID:        email.id,
			From:      email.from,
			To:        email.to,
			Subject:   email.subject,
			Body:      email.body,
			Headers:   email.headers,
			CreatedAt: email.createdAt,
		},
	})

	return nil
}

type Header struct {
	Name  string
	Value string
}

type Email struct {
	id        string
	from      string
	to        []string
	subject   string
	body      string
	headers   []Header
	createdAt time.Time
}

func (e *Email) ID() string           { return e.id }
func (e *Email) From() string         { return e.from }
func (e *Email) To() []string         { return e.to }
func (e *Email) Subject() string      { return e.subject }
func (e *Email) Body() string         { return e.body }
func (e *Email) Headers() []Header    { return e.headers }
func (e *Email) CreatedAt() time.Time { return e.createdAt }

type EmailDTO struct {
	ID        string   `validate:"required"`
	From      string   `validate:"required,email"`
	To        []string `validate:"required,dive,email"`
	Subject   string
	Body      string
	Headers   []Header
	CreatedAt time.Time `validate:"required"`
}

var validate = validator.New()

func (q *Queue) Dequeue() (*Email, bool) {
	if len(q.emails) == 0 {
		return nil, false
	}

	email := q.emails[0]
	q.emails = q.emails[1:]

	q.events = append(q.events, &DequeuedEvent{
		EmailID: email.id,
	})

	return email, true
}

type DequeuedEvent struct {
	EmailID string
}

func (q *Queue) Mailbox() domain.Mailbox { return q.mailbox }
func (q *Queue) Emails() []*Email        { return q.emails }
func (q *Queue) Events() []any           { return q.events }

type UnmarshalQueueDTO struct {
	Mailbox           domain.Mailbox
	AllowedRecipients []string
	AllowedFrom       *string
	Limit             int
	Emails            []*Email
}

func UnmarshalQueue(dto *UnmarshalQueueDTO) *Queue {
	return &Queue{
		mailbox:           dto.Mailbox,
		allowedRecipients: dto.AllowedRecipients,
		allowedFrom:       dto.AllowedFrom,
		limit:             dto.Limit,
		emails:            dto.Emails,
	}
}

func (dto *EmailDTO) ToEmail() (*Email, error) {
	if err := validate.Struct(dto); err != nil {
		return nil, err
	}
	return &Email{
		id:        dto.ID,
		from:      dto.From,
		to:        dto.To,
		subject:   dto.Subject,
		body:      dto.Body,
		headers:   dto.Headers,
		createdAt: dto.CreatedAt,
	}, nil
}

type EnqueuedEvent struct {
	Email *EmailDTO
}
