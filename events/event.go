package events

import (
	"fmt"
	"reflect"

	"github.com/DoNewsCode/core/contract"
)

// Event is a thin wrapper for events. It implements contract.Event for any interface.
type Event struct {
	body interface{}
}

// Data returns the enclosing data in the event.
func (e Event) Data() interface{} {
	return e.body
}

// Type returns the type of the event as string.
func (e Event) Type() string {
	bType := reflect.TypeOf(e.body)
	return fmt.Sprintf("%s.%s", bType.PkgPath(), bType.Name())
}

// Of wraps any struct, making it a valid contract.Event.
func Of(event interface{}) Event {
	return Event{
		body: event,
	}
}

// From implements contract.Event for a number of events. It is particularly useful
// when constructing contract.Listener's Listen function.
func From(events ...interface{}) []contract.Event {
	var out []contract.Event
	for _, evt := range events {
		out = append(out, Of(evt))
	}
	return out
}
