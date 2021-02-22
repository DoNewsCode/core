package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/DoNewsCode/core/contract"
	"github.com/stretchr/testify/assert"
)

type MockEvent struct {
	value int
}
type MockListener struct {
	events []contract.Event
	test   func(event contract.Event) error
}

func (T MockListener) Listen() []contract.Event {
	return T.events
}

func (T MockListener) Process(ctx context.Context, event contract.Event) error {
	return T.test(event)
}

func TestDispatcher(t *testing.T) {
	cases := []struct {
		name      string
		event     MockEvent
		listeners []MockListener
	}{
		{
			"one listener",
			MockEvent{},
			[]MockListener{{
				From(MockEvent{}),
				func(event contract.Event) error {
					assert.Equal(t, 0, event.Data().(MockEvent).value)
					return nil
				},
			}},
		},
		{
			"two listener",
			MockEvent{value: 2},
			[]MockListener{
				{
					From(MockEvent{}),
					func(event contract.Event) error {
						assert.Equal(t, 2, event.Data().(MockEvent).value)
						return nil
					},
				},
				{
					From(MockEvent{}),
					func(event contract.Event) error {
						assert.Equal(t, 2, event.Data().(MockEvent).value)
						return nil
					},
				},
			},
		},
		{
			"no listener",
			MockEvent{value: 2},
			[]MockListener{
				{
					From(struct{}{}),
					func(event contract.Event) error {
						assert.Equal(t, 1, 2)
						return nil
					},
				},
			},
		},
		{
			"multiple events",
			MockEvent{value: 1},
			[]MockListener{
				{
					From(struct{}{}, MockEvent{}),
					func(event contract.Event) error {
						assert.Equal(t, 1, event.Data().(MockEvent).value)
						return nil
					},
				},
			},
		},
		{
			"stop propagation",
			MockEvent{value: 2},
			[]MockListener{
				{
					From(MockEvent{}),
					func(event contract.Event) error {
						return fmt.Errorf("err!")
					},
				},
				{
					From(MockEvent{}),
					func(event contract.Event) error {
						assert.Equal(t, 2, 1)
						return nil
					},
				},
			},
		},
	}

	for _, cc := range cases {
		c := cc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dispacher := SyncDispatcher{}
			for _, listener := range c.listeners {
				dispacher.Subscribe(listener)
			}
			_ = dispacher.Dispatch(context.Background(), Of(c.event))
		})
	}
}
