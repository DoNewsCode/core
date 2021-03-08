package sagas

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type correlationIdType string

const CorrelationId correlationIdType = "CorrelationId"

// Step is a step in the Saga.
type Step struct {
	Name string
	Do   endpoint.Endpoint
	Undo func(ctx context.Context) error
}

// A Saga is a model for distributed transaction. It contains a number of statically defined steps.
// Whenever any of the step fails, the saga is rolled back.
type Saga struct {
	// Name is the name of the saga. Used for log entries.
	Name string
	// Timeout is the timeout duration of the entire saga. It this timeout has passed, the saga will be rolled back.
	Timeout time.Duration
	Steps   []*Step
}

// Result is the error type of a saga execution. It contains the DoErr: the error
// occurred during executing incremental step, and UndoErr: the errors
// encountered in the compensation steps.
type Result struct {
	DoErr   error
	UndoErr []error
}

// Error implements the error interface.
func (r *Result) Error() string {
	var builder strings.Builder
	builder.WriteString("error encountered while executing saga: ")
	builder.WriteString(r.DoErr.Error())
	if len(r.UndoErr) > 0 {
		builder.WriteString("; \n")
		builder.WriteString("additional errors encountered while rolling back: ")
		for _, err := range r.UndoErr {
			builder.WriteString(err.Error())
			builder.WriteString("; \n")
		}
	}
	return builder.String()
}
