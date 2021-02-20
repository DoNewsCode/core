// Package di is a thin wrapper around dig. See https://github.com/uber-go/dig
//
// This package is not intended for direct usage. Only use it for libraries
// written for package core.
package di

import (
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
