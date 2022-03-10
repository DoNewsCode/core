package leader

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

// ErrNotALeader is an error triggered when Resign is called but the current node is not leader.
var ErrNotALeader = errors.New("not a leader")

// Driver models a external storage that can be used for leader election.
type Driver interface {
	// Campaign starts a leader election. It should block context canceled. The status must be updated with the campaign method.
	Campaign(ctx context.Context, status *atomic.Value) error
	// Resign makes the current node a follower.
	Resign(context.Context) error
}

type Dispatcher interface {
	// Fire dispatches leader election status
	Fire(ctx context.Context, status *Status) error
}

// Election is a struct that controls the leader election. Whenever the leader
// status changed on this node, an event will be triggered. See example for how
// to listen this event.
type Election struct {
	dispatcher Dispatcher
	status     *Status
	driver     Driver
}

// NewElection returns a pointer to the newly constructed Election instance.
func NewElection(dispatcher Dispatcher, driver Driver) *Election {
	return &Election{
		dispatcher: dispatcher,
		status:     &Status{isLeader: &atomic.Value{}},
		driver:     driver,
	}
}

// Campaign starts a leader election. It will block until this node becomes a leader or context cancelled.
func (e *Election) Campaign(ctx context.Context) error {
	if err := e.driver.Campaign(ctx, e.status.isLeader); err != nil {
		return fmt.Errorf("leader election failure: %w", err)
	}
	// trigger events
	e.dispatcher.Fire(ctx, e.status)
	return nil
}

// Resign gives up the leadership.
func (e *Election) Resign(ctx context.Context) error {
	if !e.status.IsLeader() {
		return ErrNotALeader
	}
	// trigger events
	defer e.dispatcher.Fire(ctx, e.status)
	return e.driver.Resign(ctx)
}

// Status returns the current status of the election.
func (e *Election) Status() *Status {
	return e.status
}

// Status is a type that describes whether the current node is leader.
type Status struct {
	isLeader *atomic.Value
}

// IsLeader returns true if the current node is leader.
func (s Status) IsLeader() bool {
	if b, ok := s.isLeader.Load().(bool); ok {
		return b
	}
	return false
}
