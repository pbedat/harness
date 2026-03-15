package decorator

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog"
)

func ApplyCommandDecorators[C any](
	handler CommandHandler[C],
	logger zerolog.Logger,
	decorators ...Decorator[C]) CommandHandler[C] {
	var h CommandHandler[C]
	h = commandLoggingDecorator[C]{
		base:   handler,
		logger: logger,
	}
	for _, decorator := range decorators {
		h = decorator.Decorate(h)
	}

	h = commandRecoveryHandler[C]{h}

	return h
}

type DecoratorFn[C any] func(h CommandHandler[C]) CommandHandler[C]

func (fn DecoratorFn[C]) Decorate(h CommandHandler[C]) CommandHandler[C] {
	return fn(h)
}

type Decorator[C any] interface {
	Decorate(h CommandHandler[C]) CommandHandler[C]
}

type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}

type commandRecoveryHandler[C any] struct {
	next CommandHandler[C]
}

func (h commandRecoveryHandler[C]) Handle(ctx context.Context, cmd C) (err error) {
	defer func() {
		e := recover()

		switch e := e.(type) {
		case nil:
			return
		case error:
			err = e
			debug.PrintStack()
			return
		default:
			err = fmt.Errorf("recovered from: %s", err)
			debug.PrintStack()
		}
	}()

	return h.next.Handle(ctx, cmd)
}
