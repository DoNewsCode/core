package sagas

import (
	"time"
)

// LogType is a type enum that describes the types of Log.
type LogType uint

const (
	// Session type logs the occurrence of a new distributed transaction.
	Session LogType = iota
	// Do type logs an incremental action in the distributed saga step.
	Do
	// Undo type logs a compensation action in the distributed saga step.
	Undo
)

// Log is the structural Log type of the distributed saga.
type Log struct {
	ID            string
	CorrelationID string
	StartedAt     time.Time
	FinishedAt    time.Time
	LogType       LogType
	StepParam     interface{}
	StepName      string
	StepError     error
}
