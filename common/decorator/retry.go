package decorator

import (
	"context"

	"github.com/sethvargo/go-retry"
)

func Retry[C any](backoff retry.Backoff) RetryDecorator[C] {
	return RetryDecorator[C]{backoff: backoff}
}

type RetryDecorator[C any] struct {
	backoff retry.Backoff
}

// Decorate implements Decorator.
func (r RetryDecorator[C]) Decorate(h CommandHandler[C]) CommandHandler[C] {
	return retryHandler[C]{backoff: r.backoff, next: h}
}

type retryHandler[C any] struct {
	backoff retry.Backoff
	next    CommandHandler[C]
}

// Handle implements CommandHandler2.
func (r retryHandler[C]) Handle(ctx context.Context, cmd C) error {
	return retry.Do(ctx, r.backoff, func(ctx context.Context) error {
		return r.next.Handle(ctx, cmd)
	})
}

var _ Decorator[struct{}] = RetryDecorator[struct{}]{}
