package event

import (
	"context"
	"sync"

	"github.com/DoNewsCode/std/pkg/contract"
)

type Dispatcher struct {
	registry map[string][]contract.Listener
	rwLock   sync.RWMutex
}

func (d *Dispatcher) Dispatch(ctx context.Context, event contract.Event) error {
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

func (d *Dispatcher) Subscribe(listener contract.Listener) {
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	if d.registry == nil {
		d.registry = make(map[string][]contract.Listener)
	}
	for _, e := range listener.Listen() {
		d.registry[e.Type()] = append(d.registry[e.Type()], listener)
	}
}
