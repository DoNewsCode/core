package srvhttp

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsModule exposes prometheus metrics to `/metrics`. This is the standard route
// for prometheus metrics scrappers.
type MetricsModule struct{}

// ProvideHttp implements container.HttpProvider
func (m MetricsModule) ProvideHttp(router *mux.Router) {
	router.PathPrefix("/metrics").Handler(promhttp.Handler())
}
