package di

import (
	"reflect"

	"go.uber.org/dig"
)

// OptionalProvider is a struct with constructor and dig options. When
// OptionalProvider is used as the element in di.Deps, the options are applied to
// the inner dig.Container automatically.
type OptionalProvider struct {
	Constructor interface{}
	Options     []dig.ProvideOption
}

// LocationForPC sets the constructor pointer to a specified location. Use this
// Options to reduce vague debug message when constructor are made by
// reflect.makeFunc. For example:
//  LocationForPC(reflect.makeFunc(...), reflect.ValueOf(realConstructor).Pointer())
func LocationForPC(constructor interface{}, pc uintptr) interface{} {
	if op, ok := constructor.(OptionalProvider); ok {
		op.Options = append(op.Options, dig.LocationForPC(pc))
		return op
	}
	return OptionalProvider{
		Constructor: constructor,
		Options:     []dig.ProvideOption{dig.LocationForPC(pc)},
	}
}

// As constructs the instance and bind it to another interface. As means to be used as an argument to graph.Provide.
// For example:
//  As(MyConstructor, new(MyAbstractInterface))
func As(constructor interface{}, as interface{}) interface{} {
	if op, ok := constructor.(OptionalProvider); ok {
		op.Options = append(op.Options, dig.As(as))
		return op
	}
	return OptionalProvider{
		Constructor: constructor,
		Options:     []dig.ProvideOption{dig.As(as)},
	}
}

// Name constructs a named instance. Name means to be used as an argument to graph.Provide.
// For example:
//  Name(MyConstructor, "foo")
func Name(constructor interface{}, name string) interface{} {
	if op, ok := constructor.(OptionalProvider); ok {
		op.Options = append(op.Options, dig.Name(name))
		return op
	}
	return OptionalProvider{
		Constructor: constructor,
		Options:     []dig.ProvideOption{dig.Name(name)},
	}
}

// Bind binds a type to another. Useful when binding implementation to
// interfaces. The arguments should be a ptr to the binding types, rather than
// the types themselves. For example:
//  Bind(new(MyConcreteStruct), new(MyAbstractInterface))
func Bind(from interface{}, to interface{}) interface{} {
	fromTypes := []reflect.Type{reflect.TypeOf(from).Elem()}
	toTypes := []reflect.Type{reflect.TypeOf(to).Elem()}
	fnType := reflect.FuncOf(fromTypes, toTypes, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		return args
	})
	return fn.Interface()
}
