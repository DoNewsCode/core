package events

import (
	"context"

	"github.com/DoNewsCode/core/contract"
)

var _ contract.Listener = (*ListenerFunc)(nil)

// Listen creates a functional listener in one line.
func Listen(topic interface{}, callback func(ctx context.Context, payload interface{}) error) *ListenerFunc {
	return &ListenerFunc{
		topic:    topic,
		callback: callback,
	}
}

// ListenerFunc is a listener that can be constructed from one function Listen.
// It listens to the given topic and then execute the callback.
type ListenerFunc struct {
	topic    interface{}
	callback func(ctx context.Context, payload interface{}) error
}

// Listen implements contract.Listener
func (f *ListenerFunc) Listen() interface{} {
	return f.topic
}

// Process implements contract.Listener
func (f *ListenerFunc) Process(ctx context.Context, payload interface{}) error {
	return f.callback(ctx, payload)
}

type onceListener struct {
	unsub func()
	contract.Listener
}

func (o *onceListener) Process(ctx context.Context, payload interface{}) error {
	// Dispatcher is synchronous, so we don't need to lock.
	defer o.unsub()
	return o.Listener.Process(ctx, payload)
}

func (o *onceListener) Equals(listener contract.Listener) bool {
	if l, ok := o.Listener.(interface {
		Equals(anotherListener contract.Listener) bool
	}); ok {
		return l.Equals(listener)
	}
	return o.Listener == listener
}
