package sagas

import "context"

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

func (saga *Saga) AddStep(step *Step) error {
	saga.steps = append(saga.steps, step)
	return nil
}
