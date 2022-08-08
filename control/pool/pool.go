// Package pool is a generic solution for async job dispatching from web
// server. While Go natively supports async jobs by using the keyword "go", but
// this may lead to several unwanted consequences. Suppose we have a typical http handler:
//
//   func Handle(req *http.Request, resp http.ResponseWriter) {}
//
// If we dispatch async jobs using "go" like this:
//
//   func Handle(req *http.Request, resp http.ResponseWriter) {
//     go AsyncWork()
//	   resp.Write([]byte("ok"))
//   }
//
// Let's go through all the disadvantages. First, the backpressure is lost.
// There is no way to limit the maximum goroutine the handler can create. clients
// can easily flood the server. Secondly, the graceful shutdown process is
// ruined. The http server can shut down itself without losing any request, but
// the async jobs created with "go" are not protected by the server. You will
// lose all unfinished jobs once the server shuts down and program exits. lastly,
// the async job may want to access the original request context, maybe for
// tracing purposes. The request context terminates at the end of the request, so
// if you are not careful, the async jobs may be relying on a dead context.
//
// Package pool creates a goroutine worker pool at beginning of the program,
// limits the maximum concurrency for you, shuts it down at the end of the request
// without losing any async jobs, and manages the context conversion for you.
//
// Add the dependency to core:
//
//   var c *core.C = core.New()
//   c.Provide(pool.Providers())
//
// Then you can inject the pool into your http handler:
//
//   type Handler struct {
//       pool *pool.Pool
//   }
//
//   func (h *Handler) ServeHTTP(req *http.Request, resp http.ResponseWriter) {
//      pool.Go(request.Context(), AsyncWork(asyncContext))
//      resp.Write([]byte("ok"))
//   }
package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DoNewsCode/core/ctxmeta"
)

type job struct {
	fn func()
}

// Pool is an async worker pool. It can be used to dispatch the async jobs from
// web servers. See the package documentation about its advantage over creating a
// goroutine directly.
type Pool struct {
	ch          chan job
	counter     *Counter
	jobCount    int32
	workerCount int32

	wg *sync.WaitGroup

	cap         int32
	concurrency int32
	timeout     time.Duration
}

// Go dispatchers a job to the async worker pool. requestContext is the context
// from http/grpc handler, and asyncContext is the context for async job handling.
// The asyncContext contains all values from requestContext, but its cancellation has
// nothing to do with the request. If the pool has reached max concurrency, the job will
// be executed in the current goroutine. In other word, the job will be executed synchronously.
func (p *Pool) Go(requestContext context.Context, function func(asyncContext context.Context)) {
	p.incJobCount()
	p.loadWorker()
	j := job{
		fn: func() {
			function(ctxmeta.WithoutCancel(requestContext))
		},
	}
	p.ch <- j
}

func (p *Pool) WorkerCount() int32 {
	return atomic.LoadInt32(&p.workerCount)
}

func (p *Pool) incJobCount() {
	atomic.AddInt32(&p.jobCount, 1)
}

func (p *Pool) decJobCount() {
	atomic.AddInt32(&p.jobCount, -1)
}

func (p *Pool) incWorkerCount() {
	atomic.AddInt32(&p.workerCount, 1)
}

func (p *Pool) decWorkerCount() {
	atomic.AddInt32(&p.workerCount, -1)
}

func (p *Pool) needIncWorker() int32 {
	// at least one worker keepalive
	if p.WorkerCount() == 0 {
		return 1
	}

	if concurrency, jobCount := atomic.LoadInt32(&p.concurrency), atomic.LoadInt32(&p.jobCount); (concurrency == 0 || p.WorkerCount() < concurrency) && jobCount >= p.cap {
		// calculate the number of workers to be added
		return jobCount/p.cap - p.WorkerCount()
	}
	return 0
}

func (p *Pool) loadWorker() {
	v := p.needIncWorker()
	if v == 0 {
		return
	}

	for i := 0; i < int(v); i++ {
		p.wg.Add(1)
		p.incWorkerCount()

		go func() {
			timer := time.NewTimer(p.timeout)
			defer func() {
				p.decWorkerCount()
				timer.Stop()
				p.wg.Done()
			}()
			for {
				select {
				case j := <-p.ch:
					p.counter.IncAsyncJob()
					timer.Reset(p.timeout)
					j.fn()
					p.decJobCount()
				case <-timer.C:
					if p.WorkerCount() > 1 && atomic.LoadInt32(&p.jobCount)/p.cap-p.WorkerCount() < 0 {
						return
					}
					timer.Reset(p.timeout)
				}
			}

		}()
	}

	return
}
