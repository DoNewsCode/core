package srvhttp

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/handlers"
)

type ApacheLogAdapter struct {
	log.Logger
}

func (a ApacheLogAdapter) Write(p []byte) (n int, err error) {
	a.Logger.Log("msg", string(p))
	return len(p), nil
}

func MakeApacheLogMiddleware(logger log.Logger) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return handlers.LoggingHandler(ApacheLogAdapter{logger}, handler)
	}
}
