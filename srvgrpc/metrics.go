package srvgrpc

import (
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// MetricsModule exposes prometheus metrics. Here only provides a simple call,
// more complex use, please refer to github.com/grpc-ecosystem/go-grpc-prometheus
type MetricsModule struct{}

// ProvideGRPC implements container.GRPCProvider
func (m MetricsModule) ProvideGRPC(server *grpc.Server) {
	grpc_prometheus.Register(server)
}
