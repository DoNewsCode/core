package events

import (
	"context"
	"sync"
)

type entry[T any] struct {
	id int
	fn func(ctx context.Context, event T) error
}

// Event is a generic event system.
type Event[T any] struct {
	nextID    int
	mu        sync.RWMutex
	listeners []entry[T]
}

// Fire fires the event to all listeners synchronously.
func (e *Event[T]) Fire(ctx context.Context, event T) error {
	e.mu.RLock()
	listeners := make([]entry[T], len(e.listeners))
	copy(listeners, e.listeners)
	e.mu.RUnlock()

	for i := range listeners {
		err := listeners[i].fn(ctx, event)
		if err != nil {
			return err
		}
	}
	return nil
}

// On registers a listener to the event at the bottom of the listener queue.
func (e *Event[T]) On(listener func(ctx context.Context, event T) error) (unsubscribe func()) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.nextID++
	id := e.nextID
	e.listeners = append(e.listeners, entry[T]{fn: listener, id: id})
	return func() {
		e.unsubscribe(id)
	}
}

// Once subscribes the listener to the dispatcher and unsubscribe the
// listener once after the event is processed by the listener.
func (e *Event[T]) Once(listener func(ctx context.Context, event T) error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var (
		nextID int
		once   sync.Once
	)

	e.nextID++
	nextID = e.nextID

	e.listeners = append(e.listeners, entry[T]{fn: func(ctx context.Context, event T) error {
		var err error
		once.Do(func() {
			e.unsubscribe(nextID)
			err = listener(ctx, event)
		})
		return err
	}, id: nextID})
}

// unsubscribe the listener from the dispatcher. If the listener doesn't exist, ErrNotSubscribed will be returned.
// If there are multiple instance of the listener provided subscribed, only one of the will be unsubscribed.
func (e *Event[T]) unsubscribe(id int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, entry := range e.listeners {
		if entry.id == id {
			e.listeners = append(e.listeners[:i], e.listeners[i+1:]...)
		}
	}
}

// Prepend adds the listener to the beginning of the listeners queue for the
// topic it listens to. The listeners will not be deduplicated. If subscribed
// more than once, the event will be added and processed more than once.
func (e *Event[T]) Prepend(listener func(ctx context.Context, event T) error) (unsubscribe func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nextID++
	id := e.nextID
	e.listeners = append([]entry[T]{{fn: listener, id: id}}, e.listeners...)
	return func() {
		e.unsubscribe(id)
	}
}

// PrependOnce adds a one-time listener function for the event it listens to, at
// the top of the listener queue waiting for the same event. The listener will be
// unsubscribed once after the event is processed by the listener .
func (e *Event[T]) PrependOnce(listener func(ctx context.Context, event T) error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var (
		nextID int
		once   sync.Once
	)

	e.nextID++
	nextID = e.nextID

	e.listeners = append([]entry[T]{{fn: func(ctx context.Context, event T) error {
		var err error
		once.Do(func() {
			e.unsubscribe(nextID)
			err = listener(ctx, event)
		})
		return err
	}, id: e.nextID}}, e.listeners...)
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
