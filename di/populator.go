// Package di is a thin wrapper around dig. See https://github.com/uber-go/dig
//
// This package is not intended for direct usage. Only use it for libraries
// written for package core.
package di

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/DoNewsCode/core/contract"
	"go.uber.org/dig"
)

type defaultPopulator struct {
	mutex   sync.Mutex
	invoker *dig.Container
}

// Populate sets targets with values from the dependency injection container
// during application initialization. All targets must be pointers to the
// values that must be populated. Pointers to structs that embed In are
// supported, which can be used to populate multiple values in a struct.
//
// This is most helpful in unit tests: it lets tests leverage Fx's automatic
// constructor wiring to build a few structs, but then extract those structs
// for further testing.
//
// Mostly copied from uber/fx. License: https://github.com/uber-go/fx/blob/master/LICENSE
func (d *defaultPopulator) Populate(target interface{}) error {
	invokeErr := func(err error) error {
		return d.invoker.Invoke(func() error {
			return err
		})
	}
	targetTypes := make([]reflect.Type, 1)
	// Validate all targets are non-nil pointers.
	if target == nil {
		return invokeErr(fmt.Errorf("failed to Populate: target is nil"))
	}
	rt := reflect.TypeOf(target)
	if rt.Kind() != reflect.Ptr {
		return invokeErr(fmt.Errorf("failed to Populate: target is not a pointer type, got %T", rt))
	}

	targetTypes[0] = rt.Elem()

	// Build a function that looks like:
	//
	// func(t1 T1, t2 T2, ...) {
	//   *targets[0] = t1
	//   *targets[1] = t2
	//   [...]
	// }
	//
	fnType := reflect.FuncOf(targetTypes, nil, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for _, arg := range args {
			reflect.ValueOf(target).Elem().Set(arg)
		}
		return nil
	})

	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.invoker.Invoke(fn.Interface())
}

// IntoPopulator converts a *dig.Container to contract.DIPopulator.
func IntoPopulator(container *dig.Container) contract.DIPopulator {
	return &defaultPopulator{
		invoker: container,
	}
}
