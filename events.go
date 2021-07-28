package core

import (
	"net"
	"net/http"

	"google.golang.org/grpc"
)

type event string

const (
	// OnHTTPServerStart is an event triggered when the http server is ready to serve
	// traffic. At this point the module is already wired up. This event is useful to
	// register service to service discovery.
	OnHTTPServerStart event = "onHTTPServerStart"

	// OnHTTPServerShutdown is an event triggered when the http server is shutting down.
	// traffic. At this point The traffic can no longer reach the server, but the
	// database and other infrastructures are not closed yet. This event is useful
	// to unregister service to service discovery.
	OnHTTPServerShutdown event = "onHTTPServerShutdown"

	// OnGRPCServerStart is an event triggered when the grpc server is ready to serve
	// traffic. At this point the module is already wired up. This event is useful to
	// register service to service discovery.
	OnGRPCServerStart event = "onGRPCServerStart"

	// OnGRPCServerShutdown is an event triggered when the http server is shutting down.
	// traffic. At this point The traffic can no longer reach the server, but the
	// database and other infrastructures are not closed yet. This event is useful
	// to unregister service to service discovery.
	OnGRPCServerShutdown event = "onGRPCServerShutdown"
)

// OnHTTPServerStartPayload is the payload of OnHTTPServerStart
type OnHTTPServerStartPayload struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// OnHTTPServerShutdownPayload is the payload of OnHTTPServerShutdown
type OnHTTPServerShutdownPayload struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// OnGRPCServerStartPayload is the payload of OnGRPCServerStart
type OnGRPCServerStartPayload struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}

// OnGRPCServerShutdownPayload is the payload of OnGRPCServerShutdown
type OnGRPCServerShutdownPayload struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}
