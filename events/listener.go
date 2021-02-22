package events

import (
	"context"
	"github.com/DoNewsCode/core/contract"
)

// Listen creates a functional listener in one line.
func Listen(events []contract.Event, callback func(ctx context.Context, event contract.Event) error) funcListener {
	return funcListener{
		events:   events,
		callback: callback,
	}
}

type funcListener struct {
	events   []contract.Event
	callback func(ctx context.Context, event contract.Event) error
}

func (f funcListener) Listen() []contract.Event {
	return f.events
}

func (f funcListener) Process(ctx context.Context, event contract.Event) error {
	return f.callback(ctx, event)
}
