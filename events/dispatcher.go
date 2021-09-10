package events

import (
	"context"
	"errors"
	"sync"

	"github.com/DoNewsCode/core/contract"
)

var ErrNotSubscribed = errors.New("not subscribed")

var _ contract.Dispatcher = (*SyncDispatcher)(nil)

// SyncDispatcher is a contract.Dispatcher implementation that dispatches events synchronously.
// SyncDispatcher is safe for concurrent use.
type SyncDispatcher struct {
	registry map[interface{}][]contract.Listener
	rwLock   sync.RWMutex
}

// Dispatch dispatches events synchronously. If any listener returns an error,
// abort the process immediately and return that error to caller.
func (d *SyncDispatcher) Dispatch(ctx context.Context, topic interface{}, payload interface{}) error {
	d.rwLock.RLock()
	listeners, ok := d.registry[topic]
	d.rwLock.RUnlock()

	if !ok {
		return nil
	}
	for _, listener := range listeners {
		if err := listener.Process(ctx, payload); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes the listener to the dispatcher. The listeners will not be
// deduplicated. If subscribed more than once, the event will be added and
// processed more than once.
func (d *SyncDispatcher) Subscribe(listener contract.Listener) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}
	d.registry[listener.Listen()] = append(d.registry[listener.Listen()], listener)
}

// SubscribeOnce subscribes the listener to the dispatcher and unsubscribe the
// listener once after the event is processed by the listener.
func (d *SyncDispatcher) SubscribeOnce(listener contract.Listener) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}
	ol := &onceListener{Listener: listener}
	ol.unsub = func() {
		d.Unsubscribe(ol)
	}
	d.registry[listener.Listen()] = append(d.registry[listener.Listen()], ol)
}

// Prepend adds the listener to the beginning of the listeners queue for the
// topic it listens to. The listeners will not be deduplicated. If subscribed
// more than once, the event will be added and processed more than once.
func (d *SyncDispatcher) Prepend(listener contract.Listener) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}
	d.registry[listener.Listen()] = append([]contract.Listener{listener}, d.registry[listener.Listen()]...)
}

// PrependOnce adds a one-time listener function for the event it listens to, at
// the top of the listener queue waiting for the same event. The listener will be
// unsubscribed once after the event is processed by the listener .
func (d *SyncDispatcher) PrependOnce(listener contract.Listener) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}
	ol := &onceListener{Listener: listener}
	ol.unsub = func() {
		d.Unsubscribe(ol)
	}
	d.registry[listener.Listen()] = append([]contract.Listener{ol}, d.registry[listener.Listen()]...)
}

// Unsubscribe unsubscribes the listener from the dispatcher. If the listener doesn't exist, ErrNotSubscribed will be returned.
// If there are multiple instance of the listener provided subscribed, only one of the will be unsubscribed.
func (d *SyncDispatcher) Unsubscribe(listener contract.Listener) error {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}

	event := listener.Listen()
	lns := d.registry[event]

	for i := range lns {
		if ol, ok := lns[i].(interface {
			Equals(anotherListener contract.Listener) bool
		}); ok && ol.Equals(listener) {
			removeListener(&lns, i)
			d.registry[event] = lns
			return nil
		}
		if lns[i] == listener {
			removeListener(&lns, i)
			d.registry[event] = lns
			return nil
		}
	}
	return ErrNotSubscribed
}

// RemoveAllListeners removes all listeners for a given event.
func (d *SyncDispatcher) RemoveAllListeners(topic interface{}) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	delete(d.registry, topic)
}

// ListenerCount returns the number of listeners for a given event.
func (d *SyncDispatcher) ListenerCount(topic interface{}) int {
	d.rwLock.RLock()
	defer d.rwLock.RUnlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}
	return len(d.registry[topic])
}

func removeListener(lns *[]contract.Listener, i int) {
	if i+1 < len(*lns) {
		*lns = append((*lns)[0:i], (*lns)[i+1:]...)
		return
	}
	*lns = (*lns)[0:i]
}
