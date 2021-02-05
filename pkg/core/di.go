package core

import (
	"fmt"
	"os"
	"reflect"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
)

var _errType = reflect.TypeOf((*error)(nil)).Elem()

func (c *C) Provide(constructor interface{}) {
	ftype := reflect.TypeOf(constructor)
	inTypes := make([]reflect.Type, 0)
	outTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		outT := ftype.Out(i)
		if isCleanup(outT) {
			continue
		}
		outTypes = append(outTypes, outT)
	}
	for i := 0; i < ftype.NumIn(); i++ {
		inT := ftype.In(i)
		inTypes = append(inTypes, inT)
	}
	fnType := reflect.FuncOf(inTypes, outTypes, ftype.IsVariadic() /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		filteredOuts := make([]reflect.Value, 0)
		outVs := reflect.ValueOf(constructor).Call(args)
		for _, v := range outVs {
			if isCleanup(v.Type()) {
				c.AddModule(v.Interface())
				continue
			}
			filteredOuts = append(filteredOuts, v)
		}
		return filteredOuts
	})
	err := c.di.Provide(fn.Interface())
	if err != nil {
		c.Err(err)
		os.Exit(1)
	}
}

func (c *C) ProvideItself() {
	c.Provide(func() contract.Env {
		return c.Env
	})
	c.Provide(func() contract.AppName {
		return c.AppName
	})
	c.Provide(func() contract.ConfigAccessor {
		return c.ConfigAccessor
	})
	c.Provide(func() contract.ConfigRouter {
		if cc, ok := c.ConfigAccessor.(contract.ConfigRouter); ok {
			return cc
		}
		return nil
	})
	c.Provide(func() contract.ConfigWatcher {
		if cc, ok := c.ConfigAccessor.(contract.ConfigWatcher); ok {
			return cc
		}
		return nil
	})
	c.Provide(func() log.Logger {
		return c.LevelLogger
	})
	c.Provide(func() contract.Dispatcher {
		return c.Dispatcher
	})
}

func (c *C) AddModuleViaFunc(function interface{}) {
	c.Provide(function)

	ftype := reflect.TypeOf(function)
	targetTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		if isErr(ftype.Out(i)) {
			continue
		}
		if isCleanup(ftype.Out(i)) {
			continue
		}
		outT := ftype.Out(i)
		targetTypes = append(targetTypes, outT)
	}
	fnType := reflect.FuncOf(targetTypes, nil, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for _, arg := range args {
			c.AddModule(arg.Interface())
		}
		return nil
	})

	c.Invoke(fn.Interface())
}

func (c *C) Invoke(function interface{}) error {
	return c.di.Invoke(function)
}

func (c *C) Populate(targets ...interface{}) error {
	// Validate all targets are non-nil pointers.
	targetTypes := make([]reflect.Type, len(targets))
	for i, t := range targets {
		if t == nil {
			return fmt.Errorf("failed to Populate: target %v is nil", i+1)
		}
		rt := reflect.TypeOf(t)
		if rt.Kind() != reflect.Ptr {
			return fmt.Errorf("failed to Populate: target %v is not a pointer type, got %T", i+1, t)
		}

		targetTypes[i] = reflect.TypeOf(t).Elem()
	}

	fnType := reflect.FuncOf(targetTypes, nil, false)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for i, arg := range args {
			reflect.ValueOf(targets[i]).Elem().Set(arg)
		}
		return nil
	})
	return c.di.Invoke(fn.Interface())
}

func isCleanup(v reflect.Type) bool {
	if v.Kind() == reflect.Func && v.NumIn() == 0 && v.NumOut() == 0 {
		return true
	}
	return false
}

func isErr(v reflect.Type) bool {
	if v.Implements(_errType) {
		return true
	}
	return false
}
