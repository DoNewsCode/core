package srvhttp

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// DebugModule defines a http provider for container.Container. It calls pprof underneath. For instance,
// `/debug/pprof/cmdline` invokes pprof.Cmdline
type DebugModule struct{}

// ProvideHttp implements container.HttpProvider
func (d DebugModule) ProvideHTTP(router *mux.Router) {
	m := mux.NewRouter()
	m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	router.PathPrefix("/debug/").Handler(m)
}
