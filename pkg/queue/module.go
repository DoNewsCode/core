package queue

import (
	"context"
	"github.com/oklog/run"
)

// Module is the registry unit of package queue.
type Module struct {
	q QueuedDispatcher
}

// NewModule creates a new QueueModule.
func NewModule(queuedDispatcher QueuedDispatcher) *Module {
	return &Module{
		q: queuedDispatcher,
	}
}

// ProvideRunGroup implements RunProvider.
func (q *Module) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(func() error {
		return q.q.Consume(ctx)
	}, func(err error) {
		cancel()
	})
}
