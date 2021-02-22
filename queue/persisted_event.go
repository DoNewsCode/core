package queue

import "time"

// PersistedEvent represents a persisted event.
type PersistedEvent struct {
	// The UniqueId identifies each individual message. Sometimes the message can have exact same content and even
	// exact same Key. UniqueId is used to differentiate them.
	UniqueId string
	// Key is the Message type. Usually it is the string name of the event type before serialized.
	Key string
	// Value is the serialized bytes of the event.
	Value []byte
	// HandleTimeout sets the upper time limit for each run of the handler. If handleTimeout exceeds, the event will
	// be put onto the timeout queue. Note: the timeout is shared among all listeners.
	HandleTimeout time.Duration
	// Backoff sets the duration before next retry.
	Backoff time.Duration
	// Attempts denotes how many retry has been attempted. It starts from 1.
	Attempts int
	// MaxAttempts denotes the maximum number of time the handler can retry before the event is put onto
	// the failed queue.
	// By default, MaxAttempts is 1.
	MaxAttempts int
}

// Type implements contract.event. It returns the Key.
func (s *PersistedEvent) Type() string {
	return s.Key
}

// Data implements contract.event. It returns the Value.
func (s *PersistedEvent) Data() interface{} {
	return s.Value
}
