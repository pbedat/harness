package events

import (
	"context"
	"log/slog"
)

type Bus struct {
	c chan any

	listeners []Listener
}

func NewBus() *Bus {
	return &Bus{
		c: make(chan any, 100),
	}
}

func (b *Bus) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-b.c:
				for _, listener := range b.listeners {
					if err := listener(context.Background(), event); err != nil {
						// Log the error and continue with the next listener
						slog.Error("error processing event", "error", err)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

type Listener func(ctx context.Context, event any) error

func (b *Bus) Publish(ctx context.Context, event any) error {
	b.c <- event
	return nil
}

func (b *Bus) Subscribe(listener Listener) {
	b.listeners = append(b.listeners, listener)
}
