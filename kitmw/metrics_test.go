package kitmw

import (
	"context"
	"testing"

	"github.com/DoNewsCode/core/key"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type hist struct {
	observed float64
}

func (h *hist) With(labelValues ...string) metrics.Histogram {
	return h
}

func (h *hist) Observe(value float64) {
	h.observed = value
}

func TestMakeMetricsMiddleware(t *testing.T) {
	var (
		original endpoint.Endpoint
		wrapped  endpoint.Endpoint
		hist     hist
	)
	original = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return nil, errors.New("")
	}
	wrapped = MakeMetricsMiddleware(&hist, key.New())(original)
	wrapped(context.Background(), nil)
	assert.NotZero(t, hist.observed)
}
