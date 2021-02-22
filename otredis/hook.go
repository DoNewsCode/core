package otredis

import (
	"context"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
)

// hook is borrowed from https://github.com/gjbae1212/opentracing-go-redis/blob/master/hook.go
// under MIT license: https://github.com/gjbae1212/opentracing-go-redis/blob/master/LICENSE
type hook struct {
	addrs    []string
	database int
	tracer   opentracing.Tracer
}

// BeforeProcess is a hook before process.
func (h hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, newCtx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, "redis:cmd")
	ext.DBType.Set(span, "redis")
	ext.DBInstance.Set(span, strconv.Itoa(h.database))
	ext.PeerAddress.Set(span, strings.Join(h.addrs, ", "))
	ext.PeerService.Set(span, "redis")
	ext.SpanKind.Set(span, ext.SpanKindEnum("client"))
	ext.DBStatement.Set(span, strings.ToUpper(cmd.Name()))
	return newCtx, nil
}

// AfterProcess is a hook after process.
func (h hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		// if context is raised an error.
		if ctx.Err() != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.Error(ctx.Err()))
		}
		span.Finish()
	}
	return nil
}

// BeforeProcessPipeline is a hook before pipeline process.
func (h hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	span, newCtx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, "redis:pipeline:cmd")
	ext.DBType.Set(span, "redis")
	ext.DBInstance.Set(span, strconv.Itoa(h.database))
	ext.PeerAddress.Set(span, strings.Join(h.addrs, ", "))
	ext.PeerService.Set(span, "redis")
	ext.SpanKind.Set(span, ext.SpanKindEnum("client"))
	merge := make([]string, len(cmds))
	for i, cmd := range cmds {
		merge[i] = strings.ToUpper(cmd.Name())
	}
	ext.DBStatement.Set(span, strings.Join(merge, " --> "))
	return newCtx, nil
}

// BeforeProcessPipeline is a hook after pipeline process.
func (h hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		// if context is raised an error.
		if ctx.Err() != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.Error(ctx.Err()))
		}
		span.Finish()
	}
	return nil
}
