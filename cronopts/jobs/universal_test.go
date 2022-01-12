package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/dag"
	"github.com/DoNewsCode/core/internal/stub"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestUniversal_error_propagation(t *testing.T) {
	d := dag.New()
	d.AddVertex(func(ctx context.Context) error { return errors.New("!") })
	j := NewFromDAG(
		"should return error",
		d,
		WithMetrics(cronopts.NewCronJobMetrics(&stub.Histogram{}, &stub.Counter{})),
		WithTracing(opentracing.NoopTracer{}),
		WithLogs(log.NewNopLogger()),
	)
	assert.Error(t, j.Do(context.Background()))
}
