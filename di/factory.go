package di

import (
	"context"
	"sync"

	"golang.org/x/sync/singleflight"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
)

// Pair is a tuple representing a connection and a closer function
type Pair struct {
	Conn   interface{}
	Closer func()
}

// Factory is a concurrent safe, generic factory for databases and connections.
type Factory struct {
	group       *singleflight.Group
	cache       sync.Map
	constructor func(name string) (Pair, error)
	reloadOnce  sync.Once
}

// NewFactory creates a new factory.
func NewFactory(constructor func(name string) (Pair, error)) *Factory {
	return &Factory{
		constructor: constructor,
		group:       &singleflight.Group{},
	}
}

// Make creates an instance under the provided name. It an instance is already
// created and it is not nil, that instance is returned to the caller.
func (f *Factory) Make(name string) (interface{}, error) {
	var err error

	conn, err, _ := f.group.Do(name, func() (interface{}, error) {
		if slot, ok := f.cache.Load(name); ok {
			return slot.(Pair).Conn, nil
		}
		slot, err := f.constructor(name)
		if err != nil {
			return nil, err
		}
		f.cache.Store(name, slot)
		return slot.Conn, nil
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// SubscribeReloadEventFrom subscribes to the reload events from dispatcher and then notifies the di
// factory to clear its cache and shutdown all connections gracefully.
func (f *Factory) SubscribeReloadEventFrom(dispatcher contract.Dispatcher) {
	if dispatcher == nil {
		return
	}
	f.reloadOnce.Do(func() {
		dispatcher.Subscribe(events.Listen(events.From(events.OnReload{}), func(ctx context.Context, event contract.Event) error {
			f.Close()
			return nil
		}))
	})
}

// List lists created instance in the factory.
func (f *Factory) List() map[string]Pair {
	var out = make(map[string]Pair)
	f.cache.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(Pair)
		return true
	})
	return out
}

// Close closes every connection created by the factory. Connections are closed
// concurrently.
func (f *Factory) Close() {
	var wg sync.WaitGroup
	f.cache.Range(func(key, value interface{}) bool {
		defer f.cache.Delete(key)

		if value.(Pair).Closer == nil {
			return true
		}
		wg.Add(1)
		go func(value Pair) {
			value.Closer()
			wg.Done()
		}(value.(Pair))
		return true
	})
	wg.Wait()
}

// CloseConn closes a specific connection in the factory.
func (f *Factory) CloseConn(name string) {
	if value, loaded := f.cache.LoadAndDelete(name); loaded {
		if value.(Pair).Closer != nil {
			value.(Pair).Closer()
		}
	}
}
