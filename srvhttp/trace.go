package srvhttp

import "github.com/opentracing-contrib/go-stdlib/nethttp"

// Trace is an alias of nethttp.Middleware. It is recommended to use the trace
// implementation in github.com/opentracing-contrib/go-stdlib. This alias serves
// as a pointer to it.
var Trace = nethttp.Middleware
