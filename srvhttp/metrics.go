package srvhttp

import (
	"net/http"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsModule exposes prometheus metrics to `/metrics`. This is the standard route
// for prometheus metrics scrappers.
type MetricsModule struct{}

// ProvideHTTP implements container.HTTPProvider
func (m MetricsModule) ProvideHTTP(router *mux.Router) {
	router.PathPrefix("/metrics").Handler(promhttp.Handler())
}

// Metrics is a middleware for standard library http package. It records the request duration in a histogram.
func Metrics(metrics *RequestDurationSeconds) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			start := time.Now()
			defer func() {
				route := mux.CurrentRoute(request)
				if route == nil {
					metrics.Route("").Observe(time.Since(start).Seconds())
					return
				}
				path, err := route.GetPathTemplate()
				if err != nil {
					metrics.Route("").Observe(time.Since(start).Seconds())
					return
				}
				metrics.Route(path).Observe(time.Since(start).Seconds())
			}()
			handler.ServeHTTP(writer, request)
		})
	}
}

// RequestDurationSeconds is a wrapper around a histogram that measures the request latency.
type RequestDurationSeconds struct {
	// histogram is the underlying histogram of RequestDurationSeconds.
	histogram metrics.Histogram

	// labels has been set
	module  bool
	service bool
	route   bool
}

// NewRequestDurationSeconds returns a new RequestDurationSeconds.
func NewRequestDurationSeconds(histogram metrics.Histogram) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: histogram,
	}
}

// Module specifies the module label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Module(module string) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram.With("module", module),
		module:    true,
		service:   r.service,
		route:     r.route,
	}
}

// Service specifies the service label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Service(service string) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram.With("service", service),
		module:    r.module,
		service:   true,
		route:     r.route,
	}
}

// Route specifies the method label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Route(route string) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram.With("route", route),
		module:    r.module,
		service:   r.service,
		route:     true,
	}
}

// Observe records the time taken to process the request.
func (r *RequestDurationSeconds) Observe(seconds float64) {
	if !r.module {
		r.histogram = r.histogram.With("module", "")
	}
	if !r.service {
		r.histogram = r.histogram.With("service", "")
	}
	if !r.route {
		r.histogram = r.histogram.With("route", "")
	}
	r.histogram.Observe(seconds)
}
