package queue

import (
	"context"
	"reflect"
	"sync"

	"github.com/DoNewsCode/std/pkg/contract"
	stdevent "github.com/DoNewsCode/std/pkg/event"
	"github.com/pkg/errors"
)

type queueDispatcher struct {
	driver       Driver
	packer       Packer
	rwLock       sync.RWMutex
	reflectTypes map[string]reflect.Type
	base         contract.Dispatcher
}

func (d *queueDispatcher) Dispatch(ctx context.Context, event contract.Event) error {
	if _, ok := event.(*SerializedMessage); ok {
		rType := d.reflectType(event.Type())
		ptr := reflect.New(rType)
		err := d.packer.Decompress(event.Data().([]byte), ptr)
		if err != nil {
			return errors.Wrapf(err, "dispatch serialized %s failed", event.Type())
		}
		return d.base.Dispatch(ctx, stdevent.NewEvent(ptr.Elem().Interface()))
	}
	if _, ok := event.(Persistent); ok {
		data, err := d.packer.Compress(event.Data())
		if err != nil {
			return errors.Wrapf(err, "dispatch deferrable %s failed", event.Type())
		}
		msg := &SerializedMessage{
			Attempts: 1,
			Value:    data,
		}
		event.(Persistent).Decorate(msg)
		return d.driver.Push(ctx, msg, event.(Persistent).Defer())
	}
	return d.base.Dispatch(ctx, event)
}

func (d *queueDispatcher) Subscribe(listener contract.Listener) {
	d.rwLock.Lock()
	for _, e := range listener.Listen() {
		d.reflectTypes[e.Type()] = reflect.TypeOf(e.Data())
	}
	d.rwLock.Unlock()
	d.base.Subscribe(listener)
}

func (d *queueDispatcher) reflectType(typeName string) reflect.Type {
	d.rwLock.RLock()
	defer d.rwLock.RUnlock()
	return d.reflectTypes[typeName]
}

func UsePacker(packer Packer) func(*queueDispatcher) {
	return func(dispatcher *queueDispatcher) {
		dispatcher.packer = packer
	}
}

func WithQueue(dispatcher contract.Dispatcher, driver Driver, opts ...func(*queueDispatcher)) *queueDispatcher {
	qd := queueDispatcher{
		driver:       driver,
		packer:       packer{},
		rwLock:       sync.RWMutex{},
		reflectTypes: make(map[string]reflect.Type),
		base:         dispatcher,
	}
	for _, f := range opts {
		f(&qd)
	}
	return &qd
}
