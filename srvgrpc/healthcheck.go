package srvgrpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// HealthCheckModule defines a grpc provider for container.Container.
type HealthCheckModule struct{}

// ProvideGRPC implements container.GRPCProvider
func (h HealthCheckModule) ProvideGRPC(server *grpc.Server) {
	srv := health.NewServer()
	srv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, srv)
}
