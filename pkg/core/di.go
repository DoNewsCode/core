package core

import (
	"fmt"
	"os"
	"reflect"
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
				c.Register(v.Interface())
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

func (c *C) RegisterFunc(function interface{}) {
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
			c.Register(arg.Interface())
		}
		return nil
	})

	c.Invoke(fn.Interface())
}

func (c *C) Invoke(function interface{}) {
	err := c.di.Invoke(function)
	if err != nil {
		c.Err(err)
		os.Exit(1)
	}
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
