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
// ruined. The http server can shutdown itself without losing any request, but
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

	"github.com/DoNewsCode/core/ctxmeta"
)

// NewPool returns *Pool
func NewPool(manager *Manager, cap int) *Pool {
	pool := Pool{
		manager:     manager,
		concurrency: make(chan struct{}, cap),
	}
	return &pool
}

// Pool is an async worker pool. It can be used to dispatch the async jobs from
// web servers. See the package documentation about its advantage over creating a
// goroutine directly.
type Pool struct {
	manager     *Manager
	concurrency chan struct{}
}

// Go dispatchers a job to the async worker pool. requestContext is the context
// from http/grpc handler, and asyncContext is the context for async job handling.
// The asyncContext contains all values from requestContext, but its cancellation has
// nothing to do with the request. If the pool has reached max concurrency, the job will
// be executed in the current goroutine. In other word, the job will be executed synchronously.
func (p *Pool) Go(requestContext context.Context, function func(asyncContext context.Context)) {
	p.concurrency <- struct{}{}
	worker := p.manager.Get()
	fn := func() {
		defer func() {
			p.manager.Release(worker)
			<-p.concurrency
		}()
		function(ctxmeta.WithoutCancel(requestContext))
	}

	select {
	case worker.jobCh <- fn:
	case <-worker.stopCh: // only executed if manager.Run is cancelled
		fn()
	}
}

// Wait waits for all the async jobs to finish.
func (p *Pool) Wait() {
	for i := 0; i < cap(p.concurrency); i++ {
		p.concurrency <- struct{}{}
	}
}
