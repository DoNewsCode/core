package srvgrpc

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/DoNewsCode/core/internal/stub"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestRequestDurationSeconds(t *testing.T) {
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	rds = rds.Module("m").Service("s").Status(1).Route("r")
	rds.Observe(5 * time.Second)

	assert.Equal(t, 5.0, histogram.ObservedValue)
	assert.ElementsMatch(t, []string{"module", "m", "service", "s", "route", "r", "status", "1"}, histogram.LabelValues)

	f := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond)
		return nil, nil
	})
	_, _ = Metrics(rds)(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/"}, f)
	assert.GreaterOrEqual(t, 1.0, histogram.ObservedValue)
}

func TestMetrics(t *testing.T) {
	handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(2, "error")
	})
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	Metrics(rds)(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/"}, handler)
	assert.Equal(t, "2", histogram.LabelValues.Label("status"))
}

func TestMetrics_data_races(t *testing.T) {
	handler := grpc.UnaryHandler(func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	h := Metrics(rds)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			h(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/"}, handler)
			wg.Done()
		}()
	}
	wg.Wait()
}
