package srvgrpc

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// MetricsModule exposes prometheus metrics. Here only provides a simple call,
// more complex use, please refer to github.com/grpc-ecosystem/go-grpc-prometheus.
//
// Need to actively provide grpc.Server:
// 		opts := []grpc.ServerOption{
//			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
//			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
//		}
//		server = grpc.NewServer(opts...)
type MetricsModule struct{}

// ProvideGRPC implements container.GRPCProvider
func (m MetricsModule) ProvideGRPC(server *grpc.Server) {
	grpc_prometheus.Register(server)
}

// Metrics is a unary interceptor for grpc package. It records the request duration in a histogram.
func Metrics(metrics *RequestDurationSeconds) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		metrics := metrics.Route(info.FullMethod)
		defer func() {
			if s, ok := status.FromError(err); ok {
				metrics = metrics.Status(int(s.Code()))
			}
			metrics.Observe(time.Since(start))
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}

// RequestDurationSeconds is a wrapper around a histogram that measures the
// request latency. The RequestDurationSeconds exposes label setters such as
// module, service and route. If a label is set more than once, the one set last
// will take precedence.
type RequestDurationSeconds struct {
	// histogram is the underlying histogram of RequestDurationSeconds.
	histogram metrics.Histogram

	// labels
	module  string
	service string
	route   string
	status  int
}

// NewRequestDurationSeconds returns a new RequestDurationSeconds instance.
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
func (r RequestDurationSeconds) Observe(duration time.Duration) {
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
