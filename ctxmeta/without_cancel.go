package ctxmeta

import (
	"context"
	"time"
)

type asyncContext struct {
	ctx context.Context
}

func (a asyncContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (a asyncContext) Done() <-chan struct{} {
	return nil
}

func (a asyncContext) Err() error {
	return nil
}

func (a asyncContext) Value(key interface{}) interface{} {
	return a.ctx.Value(key)
}

// WithoutCancel creates a new context from an existing context and inherits all
// values from the existing context. However if the existing context is
// cancelled, timeouts or passes deadline, the returning context will not be
// affected. This is useful in an async HTTP handler. When the http response is sent,
// the request context will be cancelled. If you still want to access the value from request context (eg. tracing),
// you can use:
//  func(request *http.Request, responseWriter http.ResponseWriter) {
//    go DoSomeSlowDatabaseOperation(WithoutCancel(request.Context()))
//	  responseWriter.Write([]byte("ok"))
//  }
func WithoutCancel(requestScopeContext context.Context) (valueOnlyContext context.Context) {
	if requestScopeContext == nil {
		panic("cannot create context from nil parent")
	}
	return asyncContext{requestScopeContext}
}
