package srvhttp

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Metrics(router *mux.Router) {
	router.PathPrefix("/metrics").Handler(promhttp.Handler())
}

func MakeMetricsMiddleware() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		router := mux.NewRouter()
		router.PathPrefix("/metrics").Handler(promhttp.Handler())
		router.PathPrefix("/").Handler(handler)
		return router
	}
}
