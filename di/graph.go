// Package di is a thin wrapper around dig. See https://github.com/uber-go/dig
//
// This package is not intended for direct usage. Only use it for libraries
// written for package core.
package di

import (
	"fmt"
	"reflect"

	"github.com/DoNewsCode/core/contract"
	"go.uber.org/dig"
)

// Graph is a wrapper around dig.
type Graph struct {
	dig *dig.Container
}

// NewGraph creates a graph
func NewGraph() *Graph {
	return &Graph{dig: dig.New()}
}

// Provide teaches the container how to build values of one or more types and
// expresses their dependencies.
// The first argument of Provide is a function that accepts zero or more
// parameters and returns one or more results. The function may optionally return
// an error to indicate that it failed to build the value. This function will be
// treated as the constructor for all the types it returns. This function will be
// called AT MOST ONCE when a type produced by it, or a type that consumes this
// function's output, is requested via Invoke. If the same types are requested
// multiple times, the previously produced value will be reused. In addition to
// accepting constructors that accept dependencies as separate arguments and
// produce results as separate return values, Provide also accepts constructors
// that specify dependencies as di.In structs and/or specify results as di.Out
// structs.
func (g *Graph) Provide(constructor interface{}) error {
	return g.dig.Provide(constructor)
}

// Invoke runs the given function after instantiating its dependencies. Any
// arguments that the function has are treated as its dependencies. The
// dependencies are instantiated in an unspecified order along with any
// dependencies that they might have. The function may return an error to
// indicate failure. The error will be returned to the caller as-is.
func (g *Graph) Invoke(function interface{}) error {
	return g.dig.Invoke(function)
}

// String representation of the entire Container
func (g *Graph) String() string {
	return g.dig.String()
}

type defaultPopulater struct {
	invoker contract.DIInvoker
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
func (d *defaultPopulater) Populate(target interface{}) error {
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
	return d.invoker.Invoke(fn.Interface())
}

func IntoPopulater(container contract.DIInvoker) contract.DIPopulater {
	if populater, ok := container.(contract.DIPopulater); ok {
		return populater
	}
	return &defaultPopulater{
		invoker: container,
	}
}
