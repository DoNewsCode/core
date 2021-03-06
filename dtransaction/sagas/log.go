package sagas

import (
	"time"
)

type LogType uint

const (
	Session LogType = iota
	Executed
	Compensated
)

type Log struct {
	ID            string
	CorrelationID string
	SagaName      string
	StartedAt     time.Time
	FinishedAt    time.Time
	LogType       LogType
	StepNumber    int
	StepName      string
	StepError     error
}
