package srvhttp

import (
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestRequestDurationSeconds_data_races(t *testing.T) {
	histogram := &stub.Histogram{}
	rds := NewRequestDurationSeconds(histogram)

	f := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Millisecond)
	})
	h := Metrics(rds)(f)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			h.ServeHTTP(nil, httptest.NewRequest(http.MethodGet, "/", nil))
			wg.Done()
		}()

	}
	wg.Wait()
}
