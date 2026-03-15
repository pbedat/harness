package decorator

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/rs/zerolog"
)

func ApplyQueryDecorators[Q any, R any](handler QueryHandler[Q, R], logger zerolog.Logger,
	decorators ...QueryDecorator[Q, R]) QueryHandler[Q, R] {
	var h QueryHandler[Q, R]
	h = queryLoggingDecorator[Q, R]{
		base:   handler,
		logger: logger,
	}
	for _, decorator := range decorators {
		h = decorator.Decorate(h)
	}

	h = queryRecoveryHandler[Q, R]{h}

	return h
}

type QueryDecorator[Q, R any] interface {
	Decorate(h QueryHandler[Q, R]) QueryHandler[Q, R]
}

type QueryDecoratorFn[Q, R any] func(h QueryHandler[Q, R]) QueryHandler[Q, R]

func (fn QueryDecoratorFn[Q, R]) Decorate(h QueryHandler[Q, R]) QueryHandler[Q, R] {
	return fn(h)
}

type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

type queryRecoveryHandler[Q, R any] struct {
	next QueryHandler[Q, R]
}

func (h queryRecoveryHandler[Q, R]) Handle(ctx context.Context, q Q) (r R, err error) {
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
			err = fmt.Errorf("recovered from: %v", e)
			debug.PrintStack()
		}
	}()

	return h.next.Handle(ctx, q)
}
