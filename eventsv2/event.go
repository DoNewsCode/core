package eventsv2

import (
	"context"
	"errors"
	"sync"
)

// ErrNotSubscribed is returned when trying to unsubscribe a listener that not subscribing.
var ErrNotSubscribed = errors.New("not subscribed")

type entry[T any] struct {
	id int
	fn func(ctx context.Context, event T) error
}

type Event[T any] struct {
	nextID    int
	mu        sync.RWMutex
	listeners []entry[T]
}

func (e *Event[T]) Dispatch(ctx context.Context, event T) error {
	e.mu.RLock()
	listeners := make([]entry[T], len(e.listeners))
	copy(listeners, e.listeners)
	e.mu.RUnlock()

	for i, _ := range listeners {
		err := listeners[i].fn(ctx, event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Event[T]) Subscribe(listener func(ctx context.Context, event T) error) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.nextID++
	e.listeners = append(e.listeners, entry[T]{fn: listener, id: e.nextID})
	return e.nextID
}

// SubscribeOnce subscribes the listener to the dispatcher and unsubscribe the
// listener once after the event is processed by the listener.
func (e *Event[T]) SubscribeOnce(listener func(ctx context.Context, event T) error) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.nextID++
	nextID := e.nextID
	e.listeners = append(e.listeners, entry[T]{fn: func(ctx context.Context, event T) error {
		defer e.Unsubscribe(nextID)
		return listener(ctx, event)
	}, id: nextID})
	return nextID
}

// Unsubscribe unsubscribes the listener from the dispatcher. If the listener doesn't exist, ErrNotSubscribed will be returned.
// If there are multiple instance of the listener provided subscribed, only one of the will be unsubscribed.
func (e *Event[T]) Unsubscribe(id int) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, entry := range e.listeners {
		if entry.id == id {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
			return nil
		}
	}
	return ErrNotSubscribed
}

// Prepend adds the listener to the beginning of the listeners queue for the
// topic it listens to. The listeners will not be deduplicated. If subscribed
// more than once, the event will be added and processed more than once.
func (e *Event[T]) Prepend(listener func(ctx context.Context, event T) error) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	e.listeners = append([]entry[T]{{fn: listener, id: e.nextID}}, e.listeners...)
	return e.nextID
}

// PrependOnce adds a one-time listener function for the event it listens to, at
// the top of the listener queue waiting for the same event. The listener will be
// unsubscribed once after the event is processed by the listener .
func (e *Event[T]) PrependOnce(listener func(ctx context.Context, event T) error) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	nextID := e.nextID
	e.listeners = append([]entry[T]{{fn: func(ctx context.Context, event T) error {
		defer e.Unsubscribe(nextID)
		return listener(ctx, event)
	}, id: e.nextID}}, e.listeners...)
	return nextID
}

// RemoveAllListeners removes all listeners for a given event.
func (e *Event[T]) RemoveAllListeners() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners = nil
}

// ListenerCount returns the number of listeners for a given event.
func (e *Event[T]) ListenerCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.listeners)
}
