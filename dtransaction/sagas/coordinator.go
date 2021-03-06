package sagas

import (
	"context"
	"time"

	"github.com/rs/xid"
)

type Store interface {
	Log(ctx context.Context, log Log) error
	Ack(ctx context.Context, log Log) error
	UncommittedSteps(ctx context.Context, correlationId string) ([]Log, error)
	UncommittedSagas(ctx context.Context) ([]Log, error)
}

type Coordinator struct {
	CorrelationId string
	Saga          *Saga
	Store         Store
	doErr         error
	undoErr       []error
	aborted       bool
}

func (c *Coordinator) Execute(ctx context.Context) error {

	// start
	sagaLog := Log{
		ID:            xid.New().String(),
		CorrelationID: c.CorrelationId,
		SagaName:      c.Saga.Name,
		StartedAt:     time.Now(),
		LogType:       Session,
	}

	must(c.Store.Log(ctx, sagaLog))

	for i := 0; i < len(c.Saga.steps); i++ {
		c.execStep(ctx, i)
	}

	if c.aborted {
		return &Result{DoErr: c.doErr, UndoErr: c.undoErr}
	}

	// commit
	sagaLog.FinishedAt = time.Now()
	must(c.Store.Log(ctx, sagaLog))

	return nil
}

func (c *Coordinator) execStep(ctx context.Context, i int) {
	if c.aborted {
		return
	}
	stepLog := Log{
		ID:            xid.New().String(),
		CorrelationID: c.CorrelationId,
		SagaName:      c.Saga.Name,
		StartedAt:     time.Now(),
		LogType:       Executed,
		StepNumber:    i,
		StepName:      c.Saga.steps[i].Name,
	}
	must(c.Store.Log(ctx, stepLog))
	err := c.Saga.steps[i].Do(ctx, c.CorrelationId)

	stepLog.FinishedAt = time.Now()
	stepLog.StepError = err
	must(c.Store.Ack(ctx, stepLog))

	if err != nil {
		c.doErr = err
		c.abort(ctx)
	}

}

func (c *Coordinator) abort(ctx context.Context) {
	c.aborted = true
	steps, err := c.Store.UncommittedSteps(ctx, c.CorrelationId)
	if err != nil {
		panic(err)
	}
	for _, step := range steps {
		err := c.compensateStep(ctx, step)
		if err != nil {
			c.undoErr = append(c.undoErr, err)
		}
	}
}

func (c *Coordinator) compensateStep(ctx context.Context, step Log) error {
	compensateLog := Log{
		ID:            xid.New().String(),
		CorrelationID: c.CorrelationId,
		SagaName:      step.SagaName,
		StartedAt:     time.Now(),
		LogType:       Compensated,
		StepNumber:    step.StepNumber,
		StepName:      step.SagaName,
	}
	must(c.Store.Log(ctx, compensateLog))

	err := c.Saga.steps[step.StepNumber].Undo(ctx, c.CorrelationId)
	compensateLog.FinishedAt = time.Now()
	compensateLog.StepError = err

	must(c.Store.Ack(ctx, compensateLog))
	return err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
