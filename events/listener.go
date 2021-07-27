package events

import (
	"context"
)

// Listen creates a functional listener in one line.
func Listen(topic interface{}, callback func(ctx context.Context, event interface{}) error) funcListener {
	return funcListener{
		topic:    topic,
		callback: callback,
	}
}

type funcListener struct {
	topic    interface{}
	callback func(ctx context.Context, event interface{}) error
}

func (f funcListener) Listen() interface{} {
	return f.topic
}

func (f funcListener) Process(ctx context.Context, event interface{}) error {
	return f.callback(ctx, event)
}
