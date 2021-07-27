package events

import (
	"context"
)

// Listen creates a functional listener in one line.
func Listen(topic interface{}, callback func(ctx context.Context, event interface{}) error) ListenerFunc {
	return ListenerFunc{
		topic:    topic,
		callback: callback,
	}
}

// ListenerFunc is a listener that can be constructed from one function Listen.
// It listens to the given topic and then execute the callback.
type ListenerFunc struct {
	topic    interface{}
	callback func(ctx context.Context, event interface{}) error
}

func (f ListenerFunc) Listen() interface{} {
	return f.topic
}

func (f ListenerFunc) Process(ctx context.Context, event interface{}) error {
	return f.callback(ctx, event)
}
