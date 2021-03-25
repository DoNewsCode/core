package srvhttp

import (
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
)

// HealthCheckModule defines a http provider for container.Container.
// It uses github.com/heptiolabs/healthcheck underneath. It doesn't do much out of box other than providing liveness
// check at ``/live`` and readiness check at ``/ready``. End user should add health checking functionality by themself,
// e.g. probe if database connection pool has exhausted at readiness check.
type HealthCheckModule struct{}

// ProvideHTTP implements container.HTTPProvider
func (h HealthCheckModule) ProvideHTTP(router *mux.Router) {
	router.PathPrefix("/live").Handler(healthcheck.NewHandler())
	router.PathPrefix("/ready").Handler(healthcheck.NewHandler())
}
