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

// HTTPServerStartPayload is the payload of HTTPServerStart event
type HTTPServerStartPayload struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// HTTPServerShutdownPayload is the payload of HTTPServerShutdown event
type HTTPServerShutdownPayload struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// GRPCServerStartPayload is the payload of GRPCServerStart event
type GRPCServerStartPayload struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}

// GRPCServerShutdownPayload is the payload of GRPCServerShutdown event
type GRPCServerShutdownPayload struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}
