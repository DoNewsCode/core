package eventsv2

import (
	"context"
	"sync"
)

type Event[T any] struct {
	mu        sync.Mutex
	listeners []func(ctx context.Context, event T) error
}

func (e *Event[T]) Dispatch(ctx context.Context, event T) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, listener := range e.listeners {
		err := listener(ctx, event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Event[T]) Subscribe(listener func(ctx context.Context, event T) error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners = append(e.listeners, listener)
}
