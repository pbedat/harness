package domain

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/samber/lo"
)

//go:generate go run github.com/dmarkham/enumer -type=Mailbox -trimprefix=Mailbox -json
type Mailbox int

func (Mailbox) JSONSchema(gen *openapi3gen.Generator, schemas openapi3.Schemas) (*openapi3.SchemaRef, error) {
	return openapi3.NewSchemaRef("",
		openapi3.NewStringSchema().WithEnum(
			lo.Map(MailboxStrings(), func(s string, _ int) any {
				return s
			})...),
	), nil
}

const (
	MailboxUngültig Mailbox = iota
	MailboxInbox
	MailboxOutbox
	MailboxArchive
	MailboxSent
)
