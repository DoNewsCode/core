package ginmw

import (
	"github.com/DoNewsCode/core/key"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
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
	cases := []struct {
		name    string
		addPath bool
	}{
		{
			"addPath",
			true,
		},
		{
			"addPath",
			false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			metric := &mockMetric{}
			g := gin.New()
			g.Use(Metrics(metric, key.New("module", "foo"), c.addPath))
			g.Handle("GET", "/", func(context *gin.Context) {
				context.String(200, "%s", "ok")
			})
			req := httptest.NewRequest("GET", "/", nil)
			writer := httptest.NewRecorder()
			g.ServeHTTP(writer, req)
			assert.NotZero(t, metric.observed)
		})
	}
}
