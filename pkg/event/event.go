package event

import (
	"fmt"
	"reflect"

	"github.com/DoNewsCode/std/pkg/contract"
)

// Evt is a thin wrapper for events. It implements contract.Event for any interface.
type Evt struct {
	body interface{}
}

func (e Evt) Data() interface{} {
	return e.body
}

func (e Evt) Type() string {
	bType := reflect.TypeOf(e.body)
	return fmt.Sprintf("%s.%s", bType.PkgPath(), bType.Name())
}

// NewEvent implements contract.Event for any interface.
func NewEvent(evt interface{}) Evt {
	return Evt{
		body: evt,
	}
}

// Of implements contract.Event for a number of events. It is particularly useful
// when constructing contract.Listener's Listen function.
func Of(events ...interface{}) []contract.Event {
	var out []contract.Event
	for _, evt := range events {
		out = append(out, NewEvent(evt))
	}
	return out
}
