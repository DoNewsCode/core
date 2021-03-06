package sagas

import (
	"context"
)

type InProcessStore struct {
	logsById map[string][]Log
}

func (i *InProcessStore) Ack(ctx context.Context, log Log) error {
	logs := i.logsById[log.CorrelationID]
	for k := 0; k < len(logs); k++ {
		if logs[k].ID == log.ID {
			i.logsById[log.CorrelationID][k] = log
		}
	}
	return nil
}

func (i *InProcessStore) Log(ctx context.Context, log Log) error {
	if i.logsById == nil {
		i.logsById = make(map[string][]Log)
	}
	i.logsById[log.CorrelationID] = append(i.logsById[log.CorrelationID], log)
	return nil
}

func (i *InProcessStore) UncommittedSteps(ctx context.Context, correlationId string) ([]Log, error) {

	var (
		stepStates = make(map[int]Log)
	)

	for _, l := range i.logsById[correlationId] {
		if l.LogType == Committed {
			return []Log{}, nil
		}
		if l.LogType == Executed {
			stepStates[l.StepNumber] = l
		}
		if l.LogType == Compensated && (l.FinishedAt.IsZero()) && l.StepError == nil {
			delete(stepStates, l.StepNumber)
		}
	}
	var steps []Log
	for k := range stepStates {
		steps = append(steps, stepStates[k])
	}
	return steps, nil

}

func (i *InProcessStore) UncommittedSagas(ctx context.Context) ([]Log, error) {
	var logs []Log
	for k := range i.logsById {
		parts, err := i.UncommittedSteps(ctx, k)
		if err != nil {
			return nil, err
		}
		logs = append(logs, parts...)
	}
	return logs, nil
}
