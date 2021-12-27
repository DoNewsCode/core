package srvgrpc

import (
	"context"
	"github.com/DoNewsCode/core/internal/stub"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestRequestDurationSeconds(t *testing.T) {
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	rds = rds.Module("m").Service("s").Route("r")
	rds.Observe(5)

	assert.Equal(t, 5.0, histogram.ObservedValue)
	assert.ElementsMatch(t, []string{"module", "m", "service", "s", "route", "r"}, histogram.LabelValues)

	f := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return nil, nil
	})
	_, _ = Metrics(rds)(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/"}, f)
	assert.GreaterOrEqual(t, 1.0, histogram.ObservedValue)
}
