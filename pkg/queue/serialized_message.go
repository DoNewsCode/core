package queue

import "time"

type SerializedMessage struct {
	UniqueId      string
	Key           string
	Value         []byte
	HandleTimeout time.Duration
	Backoff       time.Duration
	Attempts      int
	MaxAttempts   int
}

func (s *SerializedMessage) Type() string {
	return s.Key
}

func (s *SerializedMessage) Data() interface{} {
	return s.Value
}
