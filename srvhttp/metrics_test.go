package srvhttp

import (
	"github.com/DoNewsCode/core/internal/stub"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestDurationSeconds(t *testing.T) {
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	rds = rds.Module("m").Service("s").Route("r")
	rds.Observe(5)

	assert.Equal(t, 5.0, histogram.ObservedValue)
	assert.ElementsMatch(t, []string{"module", "m", "service", "s", "route", "r"}, histogram.LabelValues)

	f := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Millisecond)
	})
	h := Metrics(rds)(f)
	h.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.GreaterOrEqual(t, 1.0, histogram.ObservedValue)
}

func TestRequestDurationSeconds_noPanicWhenMissingLabels(t *testing.T) {
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	rds.Observe(50)
	assert.ElementsMatch(t, []string{"module", "unknown", "service", "unknown", "route", "unknown"}, histogram.LabelValues)
}
