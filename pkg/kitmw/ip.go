package kitmw

import (
	"context"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
	"net/http"
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
	httptransport "github.com/go-kit/kit/transport/http"
)

// IpToHTTPContext is a httptransport.RequestFun that put client ip to context,
// under contract.IpKey.
func IpToHTTPContext() httptransport.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		return context.WithValue(ctx, contract.IpKey, realIP(r))
	}
}

// IpToGRPCContext is a grpctransport.ServerRequestFunc that put client ip to
// context, under contract.IpKey.
func IpToGRPCContext() grpctransport.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		remote, _ := peer.FromContext(ctx)
		return context.WithValue(ctx, contract.IpKey, remote.Addr.String())
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
