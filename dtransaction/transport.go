package dtransaction

import (
	"context"
	stdhttp "net/http"

	"github.com/go-kit/kit/transport/grpc"
	"github.com/go-kit/kit/transport/http"
	"google.golang.org/grpc/metadata"
)

const (
	header      string = "X-TX-CORRELATION-ID"
	headerHTTP2 string = "x-tx-correlation-id"
)

// HTTPToContext moves a CorrelationID from request header to context. Particularly
// useful for servers.
func HTTPToContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token := r.Header.Get(header)
		if token == "" {
			return ctx
		}
		return context.WithValue(ctx, CorrelationID, token)
	}
}

// ContextToHTTP moves a CorrelationID from context to request header. Particularly
// useful for clients.
func ContextToHTTP() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		token, ok := ctx.Value(CorrelationID).(string)
		if ok {
			r.Header.Add(header, token)
		}
		return ctx
	}
}

// GRPCToContext moves a CorrelationID from grpc metadata to context. Particularly
// userful for servers.
func GRPCToContext() grpc.ServerRequestFunc {
	return func(ctx context.Context, md metadata.MD) context.Context {
		// capital "Key" is illegal in HTTP/2.
		tokens, ok := md[headerHTTP2]
		if !ok {
			return ctx
		}
		if len(tokens) <= 0 {
			return ctx
		}
		ctx = context.WithValue(ctx, CorrelationID, tokens[len(tokens)-1])
		return ctx
	}
}

// ContextToGRPC moves a CorrelationID from context to grpc metadata. Particularly
// useful for clients.
func ContextToGRPC() grpc.ClientRequestFunc {
	return func(ctx context.Context, md *metadata.MD) context.Context {
		token, ok := ctx.Value(CorrelationID).(string)
		if ok {
			// capital "Key" is illegal in HTTP/2.
			(*md)[headerHTTP2] = []string{token}
		}
		return ctx
	}
}
