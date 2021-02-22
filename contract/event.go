package contract

import "context"

// Event is an interface for event, the unit of message.
type Event interface {
	Type() string
	Data() interface{}
}

// Dispatcher is the event registry that is able to send event to each listener.
type Dispatcher interface {
	Dispatch(ctx context.Context, event Event) error
	Subscribe(listener Listener)
}

// Listener is the handler for event.
type Listener interface {
	Listen() []Event
	Process(ctx context.Context, event Event) error
}
