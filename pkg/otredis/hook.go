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

// reference: https://github.com/opentracing/specification/blob/master/semantic_conventions.md

type Hook struct {
	addrs    []string
	database int
	tracer   opentracing.Tracer
}

func NewHook(tracer opentracing.Tracer, addrs []string, database int) Hook {
	return Hook{addrs: addrs, database: database, tracer: tracer}
}

// BeforeProcess is a hook before process.
func (h Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
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
func (h Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
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
func (h Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
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
func (h Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
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
