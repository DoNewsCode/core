package kitmw

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"google.golang.org/grpc"
)

// Interceptor is a grpc UnaryInterceptor that injects the method name into
// context so it can be consumed by Go kit gRPC middlewares. The Interceptor
// typically is added at creation time of the grpc-go server.
// Like this: `grpc.NewServer(grpc.UnaryInterceptor(kitmw.Interceptor))`
func Interceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	ctx = context.WithValue(ctx, contract.RequestUrlKey, info.FullMethod)
	return handler(ctx, req)
}
