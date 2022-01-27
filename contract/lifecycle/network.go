package lifecycle

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type HTTPServerStart interface {
	Fire(ctx context.Context, payload HTTPServerStartPayload) error
	On(func(ctx context.Context, payload HTTPServerStartPayload) error) (unsubscribe func())
}

type HTTPServerShutdown interface {
	Fire(ctx context.Context, payload HTTPServerShutdownPayload) error
	On(func(ctx context.Context, payload HTTPServerShutdownPayload) error) (unsubscribe func())
}

type GRPCServerStart interface {
	Fire(ctx context.Context, payload GRPCServerStartPayload) error
	On(func(ctx context.Context, payload GRPCServerStartPayload) error) (unsubscribe func())
}

type GRPCServerShutdown interface {
	Fire(ctx context.Context, payload GRPCServerShutdownPayload) error
	On(func(ctx context.Context, payload GRPCServerShutdownPayload) error) (unsubscribe func())
}

// HTTPServerStartPayload is the payload of OnHTTPServerStart
type HTTPServerStartPayload struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// HTTPServerShutdownPayload is the payload of OnHTTPServerShutdown
type HTTPServerShutdownPayload struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// GRPCServerStartPayload is the payload of OnGRPCServerStart
type GRPCServerStartPayload struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}

// GRPCServerShutdownPayload is the payload of OnGRPCServerShutdown
type GRPCServerShutdownPayload struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}
