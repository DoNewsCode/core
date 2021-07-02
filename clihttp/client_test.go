package clihttp

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"

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
	// the mock tracer is not concurrent safe.
	//tracer := opentracing.GlobalTracer()
	tracer := opentracing.NoopTracer{}
	client := NewClient(tracer)
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			r, _ := http.NewRequest("GET", "https://baidu.com", nil)
			_, _ = client.Do(r)
		})
	}
}

func TestClient_context(t *testing.T) {
	ctx := context.Background()
	tracer := mocktracer.New()
	span1, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "span1")
	span1.SetBaggageItem("foo", "bar")
	span1.Finish()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://baidu.com", nil)

	client := NewClient(tracer)
	_, _ = client.Do(req)
	assert.Len(t, tracer.FinishedSpans(), 2)
	assert.Equal(t, "bar", tracer.FinishedSpans()[1].BaggageItem("foo"))
}
