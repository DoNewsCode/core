package event

import (
	"context"
	"fmt"
	"testing"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/stretchr/testify/assert"
)

type TE struct {
	value int
}
type TL struct {
	events []contract.Event
	test   func(event contract.Event) error
}

func (T TL) Listen() []contract.Event {
	return T.events
}

func (T TL) Process(ctx context.Context, event contract.Event) error {
	return T.test(event)
}

func TestDispatcher(t *testing.T) {
	cases := []struct {
		name      string
		event     TE
		listeners []TL
	}{
		{
			"one listener",
			TE{},
			[]TL{{
				Of(TE{}),
				func(event contract.Event) error {
					assert.Equal(t, 0, event.Data().(TE).value)
					return nil
				},
			}},
		},
		{
			"two listener",
			TE{value: 2},
			[]TL{
				{
					Of(TE{}),
					func(event contract.Event) error {
						assert.Equal(t, 2, event.Data().(TE).value)
						return nil
					},
				},
				{
					Of(TE{}),
					func(event contract.Event) error {
						assert.Equal(t, 2, event.Data().(TE).value)
						return nil
					},
				},
			},
		},
		{
			"no listener",
			TE{value: 2},
			[]TL{
				{
					Of(struct{}{}),
					func(event contract.Event) error {
						assert.Equal(t, 1, 2)
						return nil
					},
				},
			},
		},
		{
			"multiple events",
			TE{value: 1},
			[]TL{
				{
					Of(struct{}{}, TE{}),
					func(event contract.Event) error {
						assert.Equal(t, 1, event.Data().(TE).value)
						return nil
					},
				},
			},
		},
		{
			"stop propagation",
			TE{value: 2},
			[]TL{
				{
					Of(TE{}),
					func(event contract.Event) error {
						return fmt.Errorf("err!")
					},
				},
				{
					Of(TE{}),
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
			_ = dispacher.Dispatch(context.Background(), NewEvent(c.event))
		})
	}
}
