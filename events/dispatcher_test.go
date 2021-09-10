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
	topic interface{}
	test  func(event interface{}) error
}

func (T MockListener) Listen() (topic interface{}) {
	return T.topic
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

func TestUnsubscribeDuringDispatching(t *testing.T) {
	var (
		l1Called bool
		l2Called bool
	)
	dispatcher := SyncDispatcher{}

	l1 := Listen("foo", func(ctx context.Context, event interface{}) error {
		l1Called = true
		return nil
	})
	l2 := Listen("foo", func(ctx context.Context, event interface{}) error {
		l2Called = true
		// If a user unsubscribe during proccessing a dispatched events, no dead lock should occur.
		dispatcher.Unsubscribe(l1)
		return nil
	})
	dispatcher.Subscribe(l2)
	dispatcher.Subscribe(l1)
	dispatcher.Dispatch(context.Background(), "foo", nil)
	count := dispatcher.ListenerCount("foo")
	assert.Equal(t, 1, count)
	assert.True(t, l1Called)
	assert.True(t, l2Called)
}

func TestSyncDispatcher_SubscribeAndUnsubscribe(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		process func(dispatcher *SyncDispatcher, listener contract.Listener)
		count   int
	}{{
		"subscribe once",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.SubscribeOnce(listener)
		},
		1,
	}, {
		"subscribe once but unsubscribed before execute",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.SubscribeOnce(listener)
			dispatcher.Unsubscribe(listener)
		},
		0,
	}, {
		"subscribed multiple times",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.Subscribe(listener)
			dispatcher.Subscribe(listener)
		},
		4,
	}, {
		"subscribed 2 times but unsubscribed once",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.Subscribe(listener)
			dispatcher.Subscribe(listener)
			dispatcher.Unsubscribe(listener)
		},
		2,
	}, {
		"subscribed 2 times but unsubscribed all",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.Subscribe(listener)
			dispatcher.Subscribe(listener)
			dispatcher.RemoveAllListeners(listener.Listen())
		},
		0,
	}}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var count int
			dispatcher := SyncDispatcher{}
			l := Listen("foo", func(ctx context.Context, event interface{}) error {
				count++
				return nil
			})
			c.process(&dispatcher, l)
			dispatcher.Dispatch(context.Background(), "foo", nil)
			dispatcher.Dispatch(context.Background(), "foo", nil)
			assert.Equal(t, c.count, count)
		})
	}
}

func TestSyncDispatcher_Prepend(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		process func(dispatcher *SyncDispatcher, listener contract.Listener)
		order   []int
	}{{
		"prepend",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.Prepend(listener)
		},
		[]int{2, 1, 2, 1},
	}, {
		"prepend once",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.PrependOnce(listener)
		},
		[]int{2, 1},
	}, {
		"subscribe",
		func(dispatcher *SyncDispatcher, listener contract.Listener) {
			dispatcher.Subscribe(listener)
		},
		[]int{1, 2, 1, 2},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var order []int
			dispatcher := SyncDispatcher{}
			l1 := Listen("foo", func(ctx context.Context, event interface{}) error {
				order = append(order, 1)
				return nil
			})
			l2 := Listen("foo", func(ctx context.Context, event interface{}) error {
				order = append(order, 2)
				return nil
			})
			c.process(&dispatcher, l1)
			c.process(&dispatcher, l2)
			dispatcher.Dispatch(context.Background(), "foo", nil)
			dispatcher.Dispatch(context.Background(), "foo", nil)
			assert.Equal(t, c.order, order)
		})
	}
}
