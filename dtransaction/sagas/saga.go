// Package sagas implements the orchestration based saga pattern.
// See https://microservices.io/patterns/data/saga.html
package sagas

import (
	"strings"

	"github.com/go-kit/kit/endpoint"
)

// Step is a step in the Saga.
type Step struct {
	Name string
	Do   endpoint.Endpoint
	Undo endpoint.Endpoint
}

// Result is the error type of a saga execution. It contains the DoErr: the error
// occurred during executing incremental step, and UndoErr: the errors
// encountered in the compensation steps.
type Result struct {
	DoErr   []error
	UndoErr []error
}

// Error implements the error interface.
func (r *Result) Error() string {
	var builder strings.Builder
	if r.DoErr != nil {
		builder.WriteString("errors encountered while executing saga: ")
		for _, err := range r.DoErr {
			builder.WriteString(err.Error())
			builder.WriteString("; \n")
		}
	}
	if len(r.UndoErr) > 0 {
		builder.WriteString("; \n")
		builder.WriteString("errors encountered while rolling back: ")
		for _, err := range r.UndoErr {
			builder.WriteString(err.Error())
			builder.WriteString("; \n")
		}
	}
	return builder.String()
}
