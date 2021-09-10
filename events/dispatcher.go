package events

import (
	"context"
	"errors"
	"sync"

	"github.com/DoNewsCode/core/contract"
)

var ErrNotSubscribed = errors.New("not subscribed")

// SyncDispatcher is a contract.Dispatcher implementation that dispatches events synchronously.
// SyncDispatcher is safe for concurrent use.
type SyncDispatcher struct {
	registry map[interface{}][]contract.Listener
	rwLock   sync.RWMutex
}

// Dispatch dispatches events synchronously. If any listener returns an error,
// abort the process immediately and return that error to caller.
func (d *SyncDispatcher) Dispatch(ctx context.Context, topic interface{}, event interface{}) error {
	d.rwLock.RLock()
	listeners, ok := d.registry[topic]
	d.rwLock.RUnlock()

	if !ok {
		return nil
	}
	for _, listener := range listeners {
		if err := listener.Process(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe subscribes the listener to the dispatcher.
func (d *SyncDispatcher) Subscribe(listener contract.Listener) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[interface{}][]contract.Listener)
	}
	d.registry[listener.Listen()] = append(d.registry[listener.Listen()], listener)
}

// Subscribe subscribes the listener to the dispatcher.
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
			return nil
		}
		if lns[i] == listener {
			removeListener(&lns, i)
			return nil
		}
	}
	return ErrNotSubscribed
}

func removeListener(lns *[]contract.Listener, i int) {
	if i+1 < len(*lns) {
		*lns = append((*lns)[0:i], (*lns)[i+1:]...)
	}
	*lns = (*lns)[0:i]
}
