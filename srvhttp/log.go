package srvhttp

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/handlers"
)

// ApacheLogAdapter logs HTTP requests in the Apache Common Log Format.
//
// See http://httpd.apache.org/docs/2.2/logs.html#common
type ApacheLogAdapter struct {
	log.Logger
}

// Write redirect the data stream to the underlying log.Logger
func (a ApacheLogAdapter) Write(p []byte) (n int, err error) {
	a.Logger.Log("msg", string(p))
	return len(p), nil
}

// MakeApacheLogMiddleware creates a standard HTTP middleware responsible for access logging.
func MakeApacheLogMiddleware(logger log.Logger) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return handlers.LoggingHandler(ApacheLogAdapter{logger}, handler)
	}
}
