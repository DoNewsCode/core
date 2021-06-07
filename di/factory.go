package di

import (
	"sync"
)

// Pair is a tuple representing a connection and a closer function
type Pair struct {
	Conn   interface{}
	Closer func()
}

// Factory is a concurrent safe, generic factory for databases and connections.
type Factory struct {
	mutex       sync.Mutex
	cache       map[string]Pair
	constructor func(name string) (Pair, error)
}

// NewFactory creates a new factory.
func NewFactory(constructor func(name string) (Pair, error)) *Factory {
	return &Factory{
		mutex:       sync.Mutex{},
		cache:       make(map[string]Pair),
		constructor: constructor,
	}
}

// Make creates an instance under the provided name. It an instance is already
// created and it is not nil, that instance is returned to the caller.
func (f *Factory) Make(name string) (interface{}, error) {
	var err error

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if slot, ok := f.cache[name]; ok && slot.Conn != nil {
		return slot.Conn, nil
	}

	if f.cache[name], err = f.constructor(name); err != nil {
		return nil, err
	}

	return f.cache[name].Conn, nil
}

// List lists created instance in the factory.
func (f *Factory) List() map[string]Pair {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.cache
}

// Close closes every connection created by the factory. Connections are closed
// concurrently.
func (f *Factory) Close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	var wg sync.WaitGroup
	for name := range f.cache {
		if f.cache[name].Closer == nil {
			continue
		}
		wg.Add(1)
		go func(name string) {
			f.cache[name].Closer()
			wg.Done()
		}(name)
	}
	wg.Wait()
	// Delete all. f.cache can be reused afterwards.
	for name := range f.cache {
		delete(f.cache, name)
	}
}

// CloseConn closes a specific connection in the factory.
func (f *Factory) CloseConn(name string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if pair, ok := f.cache[name]; ok && pair.Closer != nil {
		f.cache[name].Closer()
		delete(f.cache, name)
	}
}
