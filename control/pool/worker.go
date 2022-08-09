package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type worker struct {
	ch            chan job
	incWorkerChan chan int32
	jobCount      int32
	workerCount   int32

	wg sync.WaitGroup

	cap         int32
	concurrency int32
	timeout     time.Duration
}

func (w *worker) WorkerCount() int32 {
	return atomic.LoadInt32(&w.workerCount)
}

func (w *worker) incJobCount() {
	atomic.AddInt32(&w.jobCount, 1)
}

func (w *worker) decJobCount() {
	atomic.AddInt32(&w.jobCount, -1)
}

func (w *worker) incWorkerCount() {
	atomic.AddInt32(&w.workerCount, 1)
}

func (w *worker) decWorkerCount() {
	atomic.AddInt32(&w.workerCount, -1)
}

func (w *worker) run(ctx context.Context) {
	go w.incWorker(ctx)
	go w.runWorker(ctx)
}

func (w *worker) incWorker(ctx context.Context) {
	w.incWorkerChan <- 1
	timer := time.NewTimer(10 * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				timer.Reset(10 * time.Second)
				if w.WorkerCount() == 0 {
					w.incWorkerChan <- 1
				}

				if concurrency, jobCount := atomic.LoadInt32(&w.concurrency), atomic.LoadInt32(&w.jobCount); (concurrency == 0 || w.WorkerCount() < concurrency) && jobCount >= w.cap {
					// calculate the number of workers to be added
					w.incWorkerChan <- jobCount/w.cap - w.WorkerCount()
				}
			case <-ctx.Done():
				timer.Stop()
			}
		}
	}()
}

func (w *worker) runWorker(ctx context.Context) {
	for {
		select {
		case v := <-w.incWorkerChan:
			for i := 0; i < int(v); i++ {
				w.wg.Add(1)
				w.incWorkerCount()

				go func() {
					timer := time.NewTimer(w.timeout)
					defer func() {
						w.decWorkerCount()
						timer.Stop()
						w.wg.Done()
					}()
					for {
						select {
						case j := <-w.ch:
							timer.Reset(w.timeout)
							j.fn()
							w.decJobCount()
						case <-timer.C:
							if w.WorkerCount() > 1 && atomic.LoadInt32(&w.jobCount)/w.cap-w.WorkerCount() < 0 {
								return
							}
							timer.Reset(w.timeout)
						case <-ctx.Done():
							return
						}
					}

				}()
			}
		case <-ctx.Done():
			w.wg.Wait()
			return
		}
	}

}
