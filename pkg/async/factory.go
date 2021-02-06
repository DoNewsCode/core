package async

import "sync"

type Pair struct {
	Conn   interface{}
	Closer func()
}

type Factory struct {
	mutex       sync.Mutex
	cache       map[string]Pair
	constructor func(name string) (Pair, error)
}

func NewFactory(constructor func(name string) (Pair, error)) *Factory {
	return &Factory{
		mutex:       sync.Mutex{},
		cache:       make(map[string]Pair),
		constructor: constructor,
	}
}

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
}

func (f *Factory) CloseConn(name string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if pair, ok := f.cache[name]; ok && pair.Closer != nil {
		f.cache[name].Closer()
		delete(f.cache, name)
	}
}
