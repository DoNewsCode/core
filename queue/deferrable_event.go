package queue

import (
	"math/rand"
	"time"

	"github.com/DoNewsCode/core/contract"
)

// DeferrablePersistentEvent is a persisted event.
type DeferrablePersistentEvent struct {
	contract.Event
	after         time.Duration
	handleTimeout time.Duration
	maxAttempts   int
	uniqueId      string
}

// Defer defers the execution of the job for the period of time returned.
func (d DeferrablePersistentEvent) Defer() time.Duration {
	return d.after
}

// Decorate decorates the PersistedEvent of this event by adding some meta info. it is called in the QueueableDispatcher,
// after the Packer compresses the event.
func (d DeferrablePersistentEvent) Decorate(s *PersistedEvent) {
	s.UniqueId = d.uniqueId
	s.HandleTimeout = d.handleTimeout
	s.MaxAttempts = d.maxAttempts
	s.Key = d.Type()
}

// PersistOption defines some options for Persist
type PersistOption func(event *DeferrablePersistentEvent)

// Persist converts any contract.Event to DeferrablePersistentEvent. Namely, store them in external storage.
func Persist(event contract.Event, opts ...PersistOption) DeferrablePersistentEvent {
	e := DeferrablePersistentEvent{Event: event, maxAttempts: 1, handleTimeout: time.Hour, uniqueId: randomId()}
	for _, f := range opts {
		f(&e)
	}
	return e
}

// Defer is a PersistOption that defers the execution of DeferrablePersistentEvent for the period of time given.
func Defer(duration time.Duration) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.after = duration
	}
}

// ScheduleAt is a PersistOption that defers the execution of DeferrablePersistentEvent until the time given.
func ScheduleAt(t time.Time) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.after = t.Sub(time.Now())
	}
}

// Timeout is a PersistOption that defines the maximum time the event can be processed until timeout. Note: this timeout
// is shared among all listeners.
func Timeout(timeout time.Duration) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.handleTimeout = timeout
	}
}

// MaxAttempts is a PersistOption that defines how many times the event handler can be retried.
func MaxAttempts(attempts int) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.maxAttempts = attempts
	}
}

// UniqueId is a PersistOption that outsources the generation of uniqueId to the caller.
func UniqueId(id string) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.uniqueId = id
	}
}

func randomId() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 16)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
