package di

import (
	"sync"

	"golang.org/x/sync/singleflight"
)

// Pair is a tuple representing a connection and a closer function
type Pair[T any] struct {
	Conn   T
	Closer func()
}

// Factory is a concurrent safe, generic factory for connections to databases and external network services.
type Factory[T any] struct {
	group       singleflight.Group
	cache       sync.Map
	constructor func(name string) (Pair[T], error)
	reloadOnce  sync.Once
}

// NewFactory creates a new factory. The constructor is a function that is called to create a new connection.
func NewFactory[T any](constructor func(name string) (Pair[T], error)) *Factory[T] {
	return &Factory[T]{
		constructor: constructor,
	}
}

// Make returns a connection of the given name. If the connection was already
// created, it is returned from the cache. If not, a new connection will be
// established by calling the constructor.
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

// Close reloads the factory, purging all cached connections.
func (f *Factory[T]) Close() {
	f.cache.Range(func(key, value interface{}) bool {
		f.cache.Delete(key)
		pair := value.(Pair[T])
		if pair.Closer == nil {
			return true
		}
		pair.Closer()
		return true

	})
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

// CloseConn closes a specific connection in the factory.
func (f *Factory[T]) CloseConn(name string) {
	if value, loaded := f.cache.LoadAndDelete(name); loaded {
		if value.(Pair[T]).Closer != nil {
			value.(Pair[T]).Closer()
		}
	}
}
