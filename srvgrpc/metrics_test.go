package srvgrpc

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestRequestDurationSeconds(t *testing.T) {
	rds := &RequestDurationSeconds{
		Histogram: generic.NewHistogram("foo", 2),
	}
	rds = rds.Module("m").Service("s").Route("r")
	rds.Observe(5)

	assert.Equal(t, 5.0, rds.Histogram.(*generic.Histogram).Quantile(0.5))
	assert.ElementsMatch(t, []string{"module", "m", "service", "s", "route", "r"}, rds.Histogram.(*generic.Histogram).LabelValues())

	f := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return nil, nil
	})
	_, _ = Metrics(rds)(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/"}, f)
	assert.GreaterOrEqual(t, 1.0, rds.Histogram.(*generic.Histogram).Quantile(0.5))
}
