package events

import (
	"context"
	"sync"

	"github.com/DoNewsCode/core/contract"
)

// SyncDispatcher is a contract.Dispatcher implementation that dispatches events synchronously.
// SyncDispatcher is safe for concurrent use.
type SyncDispatcher struct {
	registry map[string][]contract.Listener
	rwLock   sync.RWMutex
}

// Dispatch dispatches events synchronously. If any listener returns an error,
// abort the process immediately and return that error to caller.
func (d *SyncDispatcher) Dispatch(ctx context.Context, event contract.Event) error {
	d.rwLock.RLock()
	listeners, ok := d.registry[event.Type()]
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
		d.registry = make(map[string][]contract.Listener)
	}
	for _, e := range listener.Listen() {
		d.registry[e.Type()] = append(d.registry[e.Type()], listener)
	}
}
