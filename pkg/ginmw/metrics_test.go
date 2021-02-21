package ginmw

import (
	"github.com/DoNewsCode/std/pkg/key"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockMetric struct {
	observed float64
}

func (m *mockMetric) With(labelValues ...string) metrics.Histogram {
	return m
}

func (m *mockMetric) Observe(value float64) {
	m.observed = value
}

func TestWithMetrics(t *testing.T) {
	metric := &mockMetric{}
	g := gin.New()
	g.Use(WithMetrics(metric, key.New("module", "foo"), false))
	g.Handle("GET", "/", func(context *gin.Context) {
		context.String(200, "%s", "ok")
	})
	assert.NotZero(t, metric.observed)
}
