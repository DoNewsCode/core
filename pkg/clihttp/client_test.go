package clihttp

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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
			httptest.NewRequest("GET", "https://baidu.com", nil),
			nil,
		},
		{
			"large request",
			httptest.NewRequest("POST", "https://baidu.com", strings.NewReader(strings.Repeat("t", 10))),
			[]Option{WithRequestLogThreshold(1)},
		},
		{
			"large response",
			httptest.NewRequest("GET", "https://baidu.com", nil),
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
			assert.NotZero(t, tracer.FinishedSpans())
		})
	}
}
