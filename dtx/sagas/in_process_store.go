package sagas

import (
	"context"
	"sync"
	"time"

	"github.com/DoNewsCode/core/dtx"
)

// InProcessStore creates an in process storage that implements Store.
type InProcessStore struct {
	lock         sync.Mutex
	transactions map[string][]Log
}

// NewInProcessStore creates a InProcessStore.
func NewInProcessStore() *InProcessStore {
	return &InProcessStore{
		transactions: make(map[string][]Log),
	}
}

// Ack marks the log entry as acknowledged, either with an error or not. It is
// safe to call ack to the same log entry more than once.
func (i *InProcessStore) Ack(ctx context.Context, logID string, err error) error {
	co := ctx.Value(dtx.CorrelationID).(string)
	i.lock.Lock()
	defer i.lock.Unlock()

	logs := i.transactions[co]
	for k := 0; k < len(logs); k++ {
		if logs[k].ID == logID {
			if i.transactions[co][k].LogType == Session && err == nil {
				delete(i.transactions, co)
				return nil
			}
			i.transactions[co][k].StepError = err
			i.transactions[co][k].FinishedAt = time.Now()
		}
	}
	return nil
}

// Log appends a new unacknowledged log entry to the store.
func (i *InProcessStore) Log(ctx context.Context, log Log) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.transactions == nil {
		i.transactions = make(map[string][]Log)
	}
	i.transactions[log.CorrelationID] = append(i.transactions[log.CorrelationID], log)
	return nil
}

// UnacknowledgedSteps searches the InProcessStore for unacknowledged steps under the given correlationID.
func (i *InProcessStore) UnacknowledgedSteps(ctx context.Context, correlationID string) ([]Log, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.unacknowledgedSteps(ctx, correlationID)
}

// UncommittedSagas searches the store for all uncommitted sagas, and return log entries under the matching sagas.
func (i *InProcessStore) UncommittedSagas(ctx context.Context) ([]Log, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	var logs []Log
	for k := range i.transactions {
		// For safety only. Memory store will not persist successfully finished transactions.
		if i.transactions[k][0].LogType == Session && !i.transactions[k][0].FinishedAt.IsZero() {
			return []Log{}, nil
		}

		parts, err := i.unacknowledgedSteps(ctx, k)

		if err != nil {
			return nil, err
		}
		logs = append(logs, parts...)
	}
	return logs, nil
}

func (i *InProcessStore) unacknowledgedSteps(ctx context.Context, correlationID string) ([]Log, error) {

	var (
		stepStates = make(map[string]Log)
	)

	for _, l := range i.transactions[correlationID] {
		if l.LogType == Do {
			stepStates[l.StepName] = l
		}
		if l.LogType == Undo && (!l.FinishedAt.IsZero()) && l.StepError == nil {
			delete(stepStates, l.StepName)
		}
	}
	var steps []Log
	for k := range stepStates {
		steps = append(steps, stepStates[k])
	}
	if len(steps) == 0 {
		delete(i.transactions, correlationID)
	}
	return steps, nil
}
