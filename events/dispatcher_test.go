package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockEvent struct {
	value int
}
type MockListener struct {
	topic interface{}
	test  func(event interface{}) error
}

func (T MockListener) Listen() (topic interface{}) {
	return MockEvent{}
}

func (T MockListener) Process(ctx context.Context, event interface{}) error {
	return T.test(event)
}

func TestDispatcher(t *testing.T) {
	cases := []struct {
		name      string
		topic     interface{}
		event     MockEvent
		listeners []MockListener
	}{
		{
			"one listener",
			"foo",
			MockEvent{},
			[]MockListener{{
				MockEvent{},
				func(event interface{}) error {
					assert.Equal(t, 0, event.(MockEvent).value)
					return nil
				},
			}},
		},
		{
			"two listeners",
			"foo",
			MockEvent{value: 2},
			[]MockListener{
				{
					"foo",
					func(event interface{}) error {
						assert.Equal(t, 2, event.(MockEvent).value)
						return nil
					},
				},
				{
					"foo",
					func(event interface{}) error {
						assert.Equal(t, 2, event.(MockEvent).value)
						return nil
					},
				},
			},
		},
		{
			"no listener",
			"bar",
			MockEvent{value: 2},
			[]MockListener{
				{
					"foo",
					func(event interface{}) error {
						assert.Equal(t, 1, 2)
						return nil
					},
				},
			},
		},
		{
			"multiple events",
			"foo",
			MockEvent{value: 1},
			[]MockListener{
				{
					"foo",
					func(event interface{}) error {
						assert.Equal(t, 1, event.(MockEvent).value)
						return nil
					},
				},
				{
					"bar",
					func(event interface{}) error {
						assert.Equal(t, 2, event.(MockEvent).value)
						return nil
					},
				},
			},
		},
		{
			"stop propagation",
			"foo",
			MockEvent{value: 2},
			[]MockListener{
				{
					"foo",
					func(event interface{}) error {
						return fmt.Errorf("err!")
					},
				},
				{
					"foo",
					func(event interface{}) error {
						t.Fatal("propagation should be stopped")
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
			_ = dispacher.Dispatch(context.Background(), c.topic, c.event)
		})
	}
}
