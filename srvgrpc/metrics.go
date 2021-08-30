package srvgrpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
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
		defer func() {
			metrics.Route(info.FullMethod).Observe(time.Since(start).Seconds())
		}()
		return handler(ctx, req)
	}
}

// RequestDurationSeconds is a Histogram that measures the request latency.
type RequestDurationSeconds struct {
	// Histogram is the underlying histogram of RequestDurationSeconds.
	Histogram metrics.Histogram

	// labels
	module  string
	service string
	route   string
}

// Module specifies the module label for RequestDurationSeconds.
func (r RequestDurationSeconds) Module(module string) RequestDurationSeconds {
	r.module = module
	return r
}

// Service specifies the service label for RequestDurationSeconds.
func (r RequestDurationSeconds) Service(service string) RequestDurationSeconds {
	r.service = service
	return r
}

// Route specifies the method label for RequestDurationSeconds.
func (r RequestDurationSeconds) Route(route string) RequestDurationSeconds {
	r.route = route
	return r
}

// Observe records the time taken to process the request.
func (r RequestDurationSeconds) Observe(seconds float64) {
	r.Histogram.With("module", r.module, "service", r.service, "route", r.route).Observe(seconds)
}
