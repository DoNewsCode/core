package sagas

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/dtransaction"
	"github.com/rs/xid"
)

// Store is the interface to persist logs of transactions.
type Store interface {
	Log(ctx context.Context, log Log) error
	Ack(ctx context.Context, id string, err error) error
	UnacknowledgedSteps(ctx context.Context, correlationId string) ([]Log, error)
	UncommittedSagas(ctx context.Context) ([]Log, error)
}

// Coordinator is a distributed transaction coordinator. It should be initialized
// by directly assigning its public members.
type Coordinator struct {
	correlationId string
	Saga          *Saga
	Store         Store
	doErr         error
	undoErr       []error
	aborted       bool
}

// Execute initiates a new transaction with the given parameter. Each call to
// Execute will generate a unique CorrelationId. This id will be stored both in
// the log and in the context. Upstream services can use this id to guarantee
// idempotency and issue transaction locks.
func (c *Coordinator) Execute(ctx context.Context, request interface{}) (response interface{}, err error) {
	c.correlationId = xid.New().String()
	ctx = context.WithValue(ctx, dtransaction.CorrelationID, c.correlationId)
	ctx, cancel := context.WithTimeout(ctx, c.Saga.Timeout)
	defer cancel()

	// start
	sagaLog := Log{
		ID:            xid.New().String(),
		CorrelationID: c.correlationId,
		SagaName:      c.Saga.Name,
		StartedAt:     time.Now(),
		LogType:       Session,
		StepParam:     request,
	}

	must(c.Store.Log(ctx, sagaLog))

	for i := 0; i < len(c.Saga.Steps); i++ {
		response, err = c.execStep(ctx, i, request)
		if err != nil {
			c.doErr = err
			c.abort(ctx)
			return nil, &Result{DoErr: c.doErr, UndoErr: c.undoErr}
		}
		request = response
	}

	// commit
	sagaLog.FinishedAt = time.Now()
	must(c.Store.Ack(ctx, sagaLog.ID, nil))

	return response, nil
}

func (c *Coordinator) execStep(ctx context.Context, i int, request interface{}) (response interface{}, err error) {

	logId := xid.New().String()
	stepLog := Log{
		ID:            logId,
		CorrelationID: c.correlationId,
		SagaName:      c.Saga.Name,
		StartedAt:     time.Now(),
		LogType:       Do,
		StepNumber:    i,
		StepName:      c.Saga.Steps[i].Name,
		StepParam:     request,
	}
	must(c.Store.Log(ctx, stepLog))
	response, err = c.Saga.Steps[i].Do(ctx, request)

	must(c.Store.Ack(ctx, logId, err))

	return response, err
}

func (c *Coordinator) abort(ctx context.Context) {
	steps, err := c.Store.UnacknowledgedSteps(ctx, c.correlationId)
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
	logId := xid.New().String()
	compensateLog := Log{
		ID:            logId,
		CorrelationID: c.correlationId,
		SagaName:      step.SagaName,
		StartedAt:     time.Now(),
		LogType:       Undo,
		StepNumber:    step.StepNumber,
		StepName:      step.StepName,
	}
	must(c.Store.Log(ctx, compensateLog))

	err := c.Saga.Steps[step.StepNumber].Undo(ctx, step.StepParam)

	must(c.Store.Ack(ctx, logId, err))
	return err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
