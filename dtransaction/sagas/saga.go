package sagas

import (
	"context"
	"strings"
)

type Step struct {
	Name string
	Do   func(ctx context.Context, correlationId string) error
	Undo func(ctx context.Context, correlationId string) error
}

type Saga struct {
	Name  string
	steps []*Step
}

type Result struct {
	DoErr   error
	UndoErr []error
}

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

func (saga *Saga) AddStep(step *Step) error {
	saga.steps = append(saga.steps, step)
	return nil
}
