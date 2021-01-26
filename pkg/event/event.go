package event

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/contract"
)

type Evt struct {
	body interface{}
}

func (e Evt) Data() interface{} {
	return e.body
}

func (e Evt) Type() string {
	return fmt.Sprintf("%T", e.body)
}

func NewEvent(i interface{}) Evt {
	return Evt{
		body: i,
	}
}

func Of(i ...interface{}) []contract.Event {
	var out []contract.Event
	for _, ii := range i {
		out = append(out, NewEvent(ii))
	}
	return out
}
