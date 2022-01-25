package di

import (
	"context"
	"sync"

	"golang.org/x/sync/singleflight"
)

// Pair is a tuple representing a connection and a closer function
type Pair[T any] struct {
	Conn   T
	Closer func()
}

// Factory is a concurrent safe, generic factory for databases and connections.
type Factory[T any] struct {
	group       singleflight.Group
	cache       sync.Map
	constructor func(name string) (Pair[T], error)
	reloadOnce  sync.Once
}

// NewFactory creates a new factory.
func NewFactory[T any](constructor func(name string) (Pair[T], error)) *Factory[T] {
	return &Factory[T]{
		constructor: constructor,
	}
}

// Make creates an instance under the provided name. It an instance is already
// created and it is not nil, that instance is returned to the caller.
func (f *Factory[T]) Make(name string) (T, error) {
	var err error

	conn, err, _ := f.group.Do(name, func() (interface{}, error) {
		if slot, ok := f.cache.Load(name); ok {
			return slot.(Pair[T]).Conn, nil
		}
		slot, err := f.constructor(name)
		if err != nil {
			return nil, err
		}
		f.cache.Store(name, slot)
		return slot.Conn, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return conn.(T), nil
}

// Reload reloads the factory, purging all cached connections.
func (f *Factory[T]) Reload(ctx context.Context) error {
	f.cache.Range(func(key, value interface{}) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			f.cache.Delete(key)
			pair := value.(Pair[T])
			if pair.Closer == nil {
				return true
			}
			pair.Closer()
			return true
		}
	})
	return ctx.Err()
}

// List lists created instance in the factory.
func (f *Factory[T]) List() map[string]Pair[T] {
	out := make(map[string]Pair[T])
	f.cache.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(Pair[T])
		return true
	})
	return out
}

// Close closes every connection created by the factory. Connections are closed
// concurrently.
func (f *Factory[T]) Close() {
	var wg sync.WaitGroup
	f.cache.Range(func(key, value interface{}) bool {
		defer f.cache.Delete(key)

		if value.(Pair[T]).Closer == nil {
			return true
		}
		wg.Add(1)
		go func(value Pair[T]) {
			defer wg.Done()
			value.Closer()
		}(value.(Pair[T]))
		return true
	})
	wg.Wait()
}

// CloseConn closes a specific connection in the factory.
func (f *Factory[T]) CloseConn(name string) {
	if value, loaded := f.cache.LoadAndDelete(name); loaded {
		if value.(Pair[T]).Closer != nil {
			value.(Pair[T]).Closer()
		}
	}
}
