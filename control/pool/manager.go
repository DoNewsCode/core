package pool

import (
	"context"
	"sync"
	"time"

	"github.com/DoNewsCode/core/ctxmeta"
	"github.com/oklog/run"
)

// Manager manages a pool of workers.
type Manager struct {
	workers          chan *Worker
	maxDuration      time.Duration
	startWorkerCh    chan *Worker
	managerStoppedCh chan struct{}
}

// NewManager returns a new manager.
func NewManager() *Manager {
	return &Manager{
		workers:          make(chan *Worker, 1000),
		startWorkerCh:    make(chan *Worker),
		managerStoppedCh: make(chan struct{}),
	}
}

// Get returns a worker from the free list. If the free list is empty, create a new one.
func (m *Manager) Get() *Worker {
	var w *Worker
	select {
	case w = <-m.workers:
	default:
		w = NewWorker()
		select {
		case m.startWorkerCh <- w:
		case <-m.managerStoppedCh:
			w.Stop()
		}
	}
	return w
}

// Release put the worker back into the free list. If the free list is full,
// discard the worker. If the worker has surpassed the max duration, discard and
// managerStoppedCh the worker.
func (m *Manager) Release(w *Worker) {
	if time.Now().Sub(w.startTime) > m.maxDuration {
		w.Stop()
		return
	}

	select {
	case m.workers <- w:
	default:
	}
}

// Go runs function with no concurrency limit.
func (m *Manager) Go(ctx context.Context, f func(context.Context)) {
	w := m.Get()
	ctx = ctxmeta.WithoutCancel(ctx)
	fn := func() {
		f(ctx)
		m.Release(w)
	}
	select {
	case w.jobCh <- fn:
	case <-w.stopCh: // only executed if manager.Run is cancelled
		fn()
	}
}

// Run starts the manager. It should be called during the initialization of the program.
func (m *Manager) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	for {
		select {
		case w := <-m.startWorkerCh:
			wg.Add(1)
			go func(w *Worker) {
				w.Run(ctx)
				wg.Done()
			}(w)
		case <-ctx.Done():
			close(m.managerStoppedCh)
			wg.Wait()
			return nil
		}
	}
}

// Module implements the di.Modular interface.
func (m *Manager) Module() interface{} {
	return m
}

// ProvideRunGroup implements the contract.RunProvider interface.
func (m *Manager) ProvideRunGroup(g *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	g.Add(func() error {
		return m.Run(ctx)
	}, func(err error) {
		cancel()
	})
}
