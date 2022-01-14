package srvhttp

import (
	"net/http"
	"strconv"
	"time"

	"github.com/felixge/httpsnoop"
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
			collection := httpsnoop.CaptureMetrics(handler, writer, request)
			metrics = metrics.Status(collection.Code)
			route := mux.CurrentRoute(request)
			if route == nil {
				metrics.Route("").Observe(collection.Duration)
				return
			}
			path, err := route.GetPathTemplate()
			if err != nil {
				metrics.Route("").Observe(collection.Duration)
				return
			}
			metrics.Route(path).Observe(collection.Duration)
		})
	}
}

// RequestDurationSeconds is a wrapper around a histogram that measures the
// request latency. The RequestDurationSeconds exposes label setters such as
// module, service and route. If a label is set more than once, the one set last
// will take precedence.
type RequestDurationSeconds struct {
	// histogram is the underlying histogram of RequestDurationSeconds.
	histogram metrics.Histogram

	// labels has been set
	module  string
	service string
	route   string
	status  int
}

// NewRequestDurationSeconds returns a new RequestDurationSeconds. The default
// labels are set to "unknown".
func NewRequestDurationSeconds(histogram metrics.Histogram) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: histogram,
		module:    "unknown",
		service:   "unknown",
		route:     "unknown",
		status:    0,
	}
}

// Module specifies the module label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Module(module string) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram,
		module:    module,
		service:   r.service,
		route:     r.route,
		status:    r.status,
	}
}

// Service specifies the service label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Service(service string) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram,
		module:    r.module,
		service:   service,
		route:     r.route,
		status:    r.status,
	}
}

// Route specifies the method label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Route(route string) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram,
		module:    r.module,
		service:   r.service,
		route:     route,
		status:    r.status,
	}
}

// Status specifies the status label for RequestDurationSeconds.
func (r *RequestDurationSeconds) Status(status int) *RequestDurationSeconds {
	return &RequestDurationSeconds{
		histogram: r.histogram,
		module:    r.module,
		service:   r.service,
		route:     r.route,
		status:    status,
	}
}

// Observe records the time taken to process the request.
func (r *RequestDurationSeconds) Observe(duration time.Duration) {
	r.histogram.With(
		"module",
		r.module,
		"service",
		r.service,
		"route",
		r.route,
		"status",
		strconv.Itoa(r.status),
	).Observe(duration.Seconds())
}
