package core

import (
	"google.golang.org/grpc"
	"net"
	"net/http"
)

// OnHTTPServerStart is an event triggered when the http server is ready to serve
// traffic. At this point the module is already wired up. This event is useful to
// register service to service discovery.
type OnHTTPServerStart struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// OnHTTPServerShutdown is an event triggered when the http server is shutting down.
// traffic. At this point The traffic can no longer reach the server, but the
// database and other infrastructures are not closed yet. This event is useful
// to unregister service to service discovery.
type OnHTTPServerShutdown struct {
	HTTPServer *http.Server
	Listener   net.Listener
}

// OnGRPCServerStart is an event triggered when the grpc server is ready to serve
// traffic. At this point the module is already wired up. This event is useful to
// register service to service discovery.
type OnGRPCServerStart struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}

// OnGRPCServerShutdown is an event triggered when the http server is shutting down.
// traffic. At this point The traffic can no longer reach the server, but the
// database and other infrastructures are not closed yet. This event is useful
// to unregister service to service discovery.
type OnGRPCServerShutdown struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}
