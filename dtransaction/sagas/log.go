package sagas

import (
	"time"
)

type LogType uint

const (
	Pending LogType = iota
	Executed
	Compensated
	Committed
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
