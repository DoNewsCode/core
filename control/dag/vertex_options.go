package dag

import "github.com/go-kit/log"

// VertexOption is the type of options that can be passed to the AddVertex
// function.
type VertexOption func(*vertex)

// WithName sets the name of the vertex. The name is useful for debugging.
func WithName(name string) VertexOption {
	return func(vertex *vertex) {
		vertex.name = name
	}
}

// WithLogger sets the logger for the vertex. The logger can be set to arbitrary
// log level before passing in.
func WithLogger(logger log.Logger) VertexOption {
	return func(vertex *vertex) {
		vertex.logger = logger
	}
}
