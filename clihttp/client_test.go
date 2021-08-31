package clihttp

import (
	"context"
	"io/ioutil"
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
			func() *http.Request { r, _ := http.NewRequest("GET", "https://example.com/", nil); return r }(),
			[]Option{},
		},
		{
			"large request",
			func() *http.Request {
				r, _ := http.NewRequest("POST", "https://example.com/", strings.NewReader(strings.Repeat("t", 10)))
				return r
			}(),
			[]Option{WithRequestLogThreshold(1)},
		},
		{
			"large response",
			func() *http.Request { r, _ := http.NewRequest("GET", "https://example.com/", nil); return r }(),
			[]Option{WithResponseLogThreshold(1)},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tracer := mocktracer.New()
			client := NewClient(tracer, c.Option...)
			resp, _ := client.Do(c.request)
			defer resp.Body.Close()
			assert.NotEmpty(t, tracer.FinishedSpans())
			byt, _ := ioutil.ReadAll(resp.Body)
			assert.Len(t, byt, 1256)
		})
	}
}

func TestClient_Option(t *testing.T) {
	t.Parallel()
	client := NewClient(opentracing.NoopTracer{}, []Option{
		WithResponseLogThreshold(0),
		WithRequestLogThreshold(0),
	}...)
	assert.Zero(t, client.requestLogThreshold)
	assert.Zero(t, client.responseLogThreshold)
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
}

func TestClient_race(t *testing.T) {
	// the mock tracer is not concurrent safe.
	//tracer := opentracing.GlobalTracer()
	tracer := opentracing.NoopTracer{}
	client := NewClient(tracer)
	for i := 0; i < 100; i++ {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			r, _ := http.NewRequest("GET", "https://example.com/", nil)
			resp, _ := client.Do(r)
			defer resp.Body.Close()
		})
	}
}

func TestClient_context(t *testing.T) {
	ctx := context.Background()
	tracer := mocktracer.New()
	span1, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "span1")
	span1.SetBaggageItem("foo", "bar")
	span1.Finish()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com/", nil)

	client := NewClient(tracer)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	assert.Len(t, tracer.FinishedSpans(), 2)
	assert.Equal(t, "bar", tracer.FinishedSpans()[1].BaggageItem("foo"))
}
