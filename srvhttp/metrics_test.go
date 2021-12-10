package srvhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/stretchr/testify/assert"
)

func TestRequestDurationSeconds(t *testing.T) {
	rds := &RequestDurationSeconds{
		histogram: generic.NewHistogram("foo", 2),
	}
	rds = rds.Module("m").Service("s").Route("r")
	rds.Observe(5)

	assert.Equal(t, 5.0, rds.histogram.(*generic.Histogram).Quantile(0.5))
	assert.ElementsMatch(t, []string{"module", "m", "service", "s", "route", "r"}, rds.histogram.(*generic.Histogram).LabelValues())

	f := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Millisecond)
	})
	h := Metrics(rds)(f)
	h.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.GreaterOrEqual(t, 1.0, rds.histogram.(*generic.Histogram).Quantile(0.5))
}

func TestRequestDurationSeconds_noPanicWhenMissingLabels(t *testing.T) {
	rds := &RequestDurationSeconds{
		histogram: generic.NewHistogram("foo", 2),
	}
	rds.Observe(50)
	assert.ElementsMatch(t, []string{"module", "", "service", "", "route", ""}, rds.histogram.(*generic.Histogram).LabelValues())
}
