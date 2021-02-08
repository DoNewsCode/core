package queue

import (
	"math/rand"
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
)

type Persistent interface {
	Defer() time.Duration
	Decorate(s *SerializedMessage)
}

type DeferrablePersistentEvent struct {
	contract.Event
	after         time.Duration
	handleTimeout time.Duration
	maxAttempts   int
	uniqueId      string
}

func (d DeferrablePersistentEvent) Defer() time.Duration {
	return d.after
}

func (d DeferrablePersistentEvent) Decorate(s *SerializedMessage) {
	s.UniqueId = d.uniqueId
	s.HandleTimeout = d.handleTimeout
	s.MaxAttempts = d.maxAttempts
	s.Key = d.Type()
}

type PersistOption func(event *DeferrablePersistentEvent)

func Persist(event contract.Event, opts ...PersistOption) DeferrablePersistentEvent {
	e := DeferrablePersistentEvent{Event: event, maxAttempts: 1, handleTimeout: time.Hour, uniqueId: randomId()}
	for _, f := range opts {
		f(&e)
	}
	return e
}

func Defer(duration time.Duration) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.after = duration
	}
}

func ScheduleAt(t time.Time) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.after = t.Sub(time.Now())
	}
}

func Timeout(timeout time.Duration) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.handleTimeout = timeout
	}
}

func MaxAttempts(attempts int) PersistOption {
	return func(event *DeferrablePersistentEvent) {
		event.maxAttempts = attempts
	}
}

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
