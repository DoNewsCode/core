package async

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

func Detach(ctx context.Context, callback func(ctx context.Context)) {
	span := opentracing.SpanFromContext(ctx)
	go func() {
		ctx := opentracing.ContextWithSpan(context.Background(), span)
		callback(ctx)
	}()
}
