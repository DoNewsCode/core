package leader

import (
	"context"
	"errors"
	"fmt"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"go.uber.org/atomic"
)

var ErrNotALeader = errors.New("not a leader")

// Driver models a external storage that can be used for leader election.
type Driver interface {
	// Campaign starts a leader election. It should block until elected or context canceled.
	Campaign(ctx context.Context) error
	// Resign makes the current node a follower.
	Resign(context.Context) error
}

type Election struct {
	dispatcher contract.Dispatcher
	status     *Status
	driver     Driver
}

func (e *Election) Campaign(ctx context.Context) error {
	if err := e.driver.Campaign(ctx); err != nil {
		return fmt.Errorf("not elected: %w", err)
	}
	e.status.isLeader.Store(true)
	// trigger events
	e.dispatcher.Dispatch(ctx, events.Of(e.status))
	return nil
}

func (e *Election) Resign(ctx context.Context) error {
	if e.status.isLeader.Load() != true {
		return ErrNotALeader
	}
	// trigger events
	e.dispatcher.Dispatch(ctx, events.Of(e.status))
	return e.driver.Resign(ctx)
}

// Status returns the current status of the election.
func (e *Election) Status() *Status {
	return e.status
}

type Status struct {
	isLeader *atomic.Bool
}

func (s Status) IsLeader() bool {
	return s.isLeader.Load()
}
