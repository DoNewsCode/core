package kitmw

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/go-kit/kit/endpoint"
	stdtracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// MakeLabeledTraceServerMiddleware returns a LabeledMiddleware that wraps the `next` Endpoint in an
// OpenTracing Span. The name of the operation is defined by contract.Keyer.
func MakeLabeledTraceServerMiddleware(tracer stdtracing.Tracer, keyer contract.Keyer) LabeledMiddleware {
	return func(method string, endpoint endpoint.Endpoint) endpoint.Endpoint {
		name := key.KeepOdd(keyer).Key(".", "method", method)
		return TraceConsumer(tracer, name, ext.SpanKindRPCServerEnum)(endpoint)
	}
}

// MakeTraceServerMiddleware returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span. The name of the operation is defined by contract.Keyer.
func MakeTraceServerMiddleware(tracer stdtracing.Tracer, keyer contract.Keyer) endpoint.Middleware {
	return func(endpoint endpoint.Endpoint) endpoint.Endpoint {
		name := key.KeepOdd(keyer).Key(".")
		return TraceConsumer(tracer, name, ext.SpanKindRPCServerEnum)(endpoint)
	}
}

// TraceConsumer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
//
// If `ctx` already has a Span, it is re-used and the operation name is
// overwritten. If `ctx` does not yet have a Span, one is created here.
func TraceConsumer(tracer stdtracing.Tracer, operationName string, kind ext.SpanKindEnum) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			serverSpan := stdtracing.SpanFromContext(ctx)
			if serverSpan == nil {
				// All we can do is create a new root span.
				serverSpan = tracer.StartSpan(operationName)
			} else {
				serverSpan.SetOperationName(operationName)
			}
			defer serverSpan.Finish()
			ext.SpanKind.Set(serverSpan, kind)
			if tenant, ok := ctx.Value(contract.TenantKey).(contract.Tenant); ok {
				for k, v := range tenant.KV() {
					serverSpan.SetTag(k, v)
				}
			}

			transport, _ := ctx.Value(contract.TransportKey).(string)
			serverSpan.SetTag("transport", transport)
			requestUrl, _ := ctx.Value(contract.RequestUrlKey).(string)
			serverSpan.SetTag("request.url", requestUrl)

			ctx = stdtracing.ContextWithSpan(ctx, serverSpan)
			resp, err := next(ctx, request)
			if err != nil {
				ext.Error.Set(serverSpan, true)
			}

			return resp, err
		}
	}
}

// TraceProducer returns a Middleware that wraps the `next` Endpoint in an
// OpenTracing Span called `operationName`.
func TraceProducer(tracer stdtracing.Tracer, operationName string, kind ext.SpanKindEnum) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			var clientSpan stdtracing.Span
			if parentSpan := stdtracing.SpanFromContext(ctx); parentSpan != nil {
				clientSpan = tracer.StartSpan(
					operationName,
					stdtracing.ChildOf(parentSpan.Context()),
				)
			} else {
				clientSpan = tracer.StartSpan(operationName)
			}
			defer clientSpan.Finish()
			ext.SpanKind.Set(clientSpan, kind)
			ctx = stdtracing.ContextWithSpan(ctx, clientSpan)
			resp, err := next(ctx, request)
			if err != nil {
				ext.Error.Set(clientSpan, true)
			}
			return resp, err
		}
	}
}
