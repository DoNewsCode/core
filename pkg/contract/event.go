package contract

import "context"

type Event interface {
	Type() string
	Data() interface{}
}

type Dispatcher interface {
	Dispatch(ctx context.Context, event Event) error
	Subscribe(listener Listener)
}

type Listener interface {
	Listen() []Event
	Process(ctx context.Context, event Event) error
}
