package kitmw

import (
	"context"
	"net"
	"net/http"
	"strings"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/DoNewsCode/core/contract"
	httptransport "github.com/go-kit/kit/transport/http"
)

// IPToHTTPContext is a httptransport.RequestFun that put client ip to context,
// under contract.IpKey.
func IPToHTTPContext() httptransport.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		return context.WithValue(ctx, contract.IpKey, realIP(r))
	}
}

// IPToGRPCContext is a grpctransport.ServerRequestFunc that put client ip to
// context, under contract.IpKey.
func IPToGRPCContext() grpctransport.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		remote, _ := peer.FromContext(ctx)
		ip, _, _ := net.SplitHostPort(remote.Addr.String())
		return context.WithValue(ctx, contract.IpKey, ip)
	}
}

// TransportToHTTPContext adds transport information to the context.
func TransportToHTTPContext() httptransport.RequestFunc {
	return func(ctx context.Context, request *http.Request) context.Context {
		context.WithValue(ctx, contract.RequestUrlKey, request.RequestURI)
		return context.WithValue(ctx, contract.TransportKey, HTTPKind)
	}
}

// RequestURLToHTTPContext adds path information to the context.
func RequestURLToHTTPContext() httptransport.RequestFunc {
	return func(ctx context.Context, request *http.Request) context.Context {
		return context.WithValue(ctx, contract.RequestUrlKey, request.RequestURI)
	}
}

// TransportToGRPCContext adds path and transport information to the context.
func TransportToGRPCContext() grpctransport.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		return context.WithValue(ctx, contract.TransportKey, GRPCKind)
	}
}

func realIP(r *http.Request) string {
	var ip string
	var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
		return ip
	}

	if xrip := r.Header.Get(xRealIP); xrip != "" {
		return xrip
	}

	ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	return ip
}
