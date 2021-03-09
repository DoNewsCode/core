package clihttp

import (
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
)

func TestClient_Do(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		request *http.Request
		Option  []Option
	}{
		{
			"normal",
			func() *http.Request { r, _ := http.NewRequest("GET", "https://baidu.com", nil); return r }(),
			[]Option{},
		},
		{
			"large request",
			func() *http.Request {
				r, _ := http.NewRequest("POST", "https://baidu.com", strings.NewReader(strings.Repeat("t", 10)))
				return r
			}(),
			[]Option{WithRequestLogThreshold(1)},
		},
		{
			"large response",
			func() *http.Request { r, _ := http.NewRequest("GET", "https://baidu.com", nil); return r }(),
			[]Option{WithResponseLogThreshold(1)},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tracer := mocktracer.New()
			client := NewClient(tracer, c.Option...)
			_, _ = client.Do(c.request)
			assert.NotEmpty(t, tracer.FinishedSpans())
		})
	}
}

func TestClient_race(t *testing.T) {
	cases := []struct {
		name    string
		request *http.Request
		Option  []Option
	}{
		{
			"normal",
			func() *http.Request { r, _ := http.NewRequest("GET", "https://baidu.com", nil); return r }(),
			[]Option{},
		},
	}
	for _, c := range cases {
		c := c
		// the mock tracer is not concurrent safe.
		tracer := opentracing.GlobalTracer()
		client := NewClient(tracer, c.Option...)
		for i := 0; i < 10; i++ {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				_, _ = client.Do(c.request)
			})
		}
	}
}
