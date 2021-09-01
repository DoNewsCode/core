package srvgrpc

import "github.com/opentracing-contrib/go-grpc"

// Trace is an alias of otgrpc.OpenTracingServerInterceptor. It is recommended to use the trace
// implementation in github.com/opentracing-contrib/go-grpc. This alias serves
// as a pointer to it.
var Trace = otgrpc.OpenTracingServerInterceptor
