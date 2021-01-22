package srvhttp

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
)

func MakeHealthCheckMiddleware() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		router := mux.NewRouter()
		router.PathPrefix("/live").Handler(healthcheck.NewHandler())
		router.PathPrefix("/ready").Handler(healthcheck.NewHandler())
		router.PathPrefix("/").Handler(handler)
		return router
	}
}

func HealthCheck(router *mux.Router) {
	router.PathPrefix("/live").Handler(healthcheck.NewHandler())
	router.PathPrefix("/ready").Handler(healthcheck.NewHandler())
}
