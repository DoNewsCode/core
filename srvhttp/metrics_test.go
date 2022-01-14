package srvhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoNewsCode/core/internal/stub"

	"github.com/stretchr/testify/assert"
)

func TestRequestDurationSeconds(t *testing.T) {
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)
	rds = rds.Module("m").Service("s").Route("r").Status(8)
	rds.Observe(5 * time.Second)

	assert.Equal(t, 5.0, histogram.ObservedValue)
	assert.ElementsMatch(t, []string{"module", "m", "service", "s", "route", "r", "status", "8"}, histogram.LabelValues)

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
	assert.ElementsMatch(t, []string{"module", "unknown", "service", "unknown", "route", "unknown", "status", "0"}, histogram.LabelValues)
}
