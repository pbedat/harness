package email

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pbedat/harness/modules/email/domain"
)

var validate = validator.New()

func Erstellen(dto *NewEmailDTO) (*Email, error) {
	email, err := dto.ToEmail()
	if err != nil {
		return nil, err
	}

	email.events = append(email.events, ErstelltEvent{Email: *dto})
	return email, nil
}

type NewEmailDTO struct {
	ID        string         `validate:"required"`
	Mailbox   domain.Mailbox `validate:"required"`
	From      string         `validate:"required,email"`
	To        []string       `validate:"required,dive,email"`
	Subject   string
	Body      string
	HtmlBody  string
	Headers   []Header
	CreatedAt time.Time `validate:"required"`
}

func (dto *NewEmailDTO) ToEmail() (*Email, error) {
	if err := validate.Struct(dto); err != nil {
		return nil, err
	}

	return &Email{
		id:        dto.ID,
		mailbox:   dto.Mailbox,
		from:      dto.From,
		to:        dto.To,
		subject:   dto.Subject,
		body:      dto.Body,
		htmlBody:  dto.HtmlBody,
		headers:   dto.Headers,
		createdAt: dto.CreatedAt,
	}, nil
}

func (e *Email) MarkAsRead() {
	if e.readAt != nil {
		// Already marked as read, do nothing
		return
	}

	now := time.Now()
	e.readAt = &now
}

func (e *Email) Move(newMailbox domain.Mailbox) error {

	e.mailbox = newMailbox
	e.events = append(e.events, &MovedEvent{
		EmailID:  e.id,
		Mailbox:  newMailbox,
		To:       e.to,
		From:     e.from,
		Subject:  e.subject,
		Body:     e.body,
		HtmlBody: e.htmlBody,
		Headers:  e.headers,
	})
	return nil
}

type MovedEvent struct {
	EmailID  string
	Mailbox  domain.Mailbox
	To       []string
	From     string
	Subject  string
	Body     string
	HtmlBody string
	Headers  []Header
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
	htmlBody  string
	headers   []Header
	createdAt time.Time
	readAt    *time.Time
	mailbox   domain.Mailbox

	events []any
}

type ErstelltEvent struct {
	Email NewEmailDTO
}

func (e *Email) ID() string              { return e.id }
func (e *Email) From() string            { return e.from }
func (e *Email) To() []string            { return e.to }
func (e *Email) Subject() string         { return e.subject }
func (e *Email) Body() string            { return e.body }
func (e *Email) HtmlBody() string        { return e.htmlBody }
func (e *Email) Headers() []Header       { return e.headers }
func (e *Email) CreatedAt() time.Time    { return e.createdAt }
func (e *Email) ReadAt() *time.Time      { return e.readAt }
func (e *Email) Mailbox() domain.Mailbox { return e.mailbox }
func (e *Email) Events() []any           { return e.events }

type UnmarshalEmailDTO struct {
	ID        string
	From      string
	To        []string
	Subject   string
	Body      string
	HtmlBody  string
	Headers   []Header
	CreatedAt time.Time
	Mailbox   domain.Mailbox
	ReadAt    *time.Time
}

func UnmarshalEmail(dto *UnmarshalEmailDTO) *Email {
	return &Email{
		id:        dto.ID,
		from:      dto.From,
		to:        dto.To,
		subject:   dto.Subject,
		body:      dto.Body,
		htmlBody:  dto.HtmlBody,
		headers:   dto.Headers,
		createdAt: dto.CreatedAt,
		mailbox:   dto.Mailbox,
		readAt:    dto.ReadAt,
	}
}
