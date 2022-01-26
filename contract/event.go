package contract

import "context"

// Dispatcher is the event registry that is able to send payload to each listener.
type Dispatcher interface {
	Dispatch(ctx context.Context, topic interface{}, payload interface{}) error
	Subscribe(listener Listener)
}

// Listener is the handler for event.
type Listener interface {
	Listen() (topic interface{})
	Process(ctx context.Context, payload interface{}) error
}

type ConfigReloadDispatcher interface {
	Dispatch(ctx context.Context, Config ConfigUnmarshaler) error
	Subscribe(func(ctx context.Context, Config ConfigUnmarshaler) error) int
}
