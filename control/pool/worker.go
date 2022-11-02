package pool

import (
	"context"
	"sync"
	"time"
)

type Worker struct {
	jobCh     chan func()
	startTime time.Time
	stopCh    chan struct{}
	once      sync.Once
}

func NewWorker() *Worker {
	return &Worker{
		jobCh:  make(chan func()),
		stopCh: make(chan struct{}),
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.startTime = time.Now()
	for {
		select {
		case fn := <-w.jobCh:
			fn()
		case <-ctx.Done():
			w.Stop()
			return
		case <-w.stopCh:
			return
		}
	}
}

func (w *Worker) Stop() {
	w.once.Do(func() {
		close(w.stopCh)
	})
}
