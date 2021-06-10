package srvhttp

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// DebugModule defines a http provider for container.Container. It calls pprof underneath. For instance,
// `/debug/pprof/cmdline` invokes pprof.Cmdline
type DebugModule struct{}

// ProvideHTTP implements container.HTTPProvider
func (d DebugModule) ProvideHTTP(router *mux.Router) {
	m := mux.NewRouter()
	m.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/debug/pprof/trace", pprof.Trace)
	m.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)

	router.PathPrefix("/debug/").Handler(m)
}
