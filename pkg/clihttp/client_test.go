package clihttp

import (
	"net/http"
	"strings"
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
)

type MockClient struct{}

func (m MockClient) Do(request *http.Request) (*http.Response, error) { return &http.Response{}, nil }

func TestClient_Do(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		tracer := mocktracer.New()
		client := NewClient(tracer)
		req, _ := http.NewRequest("GET", "https://baidu.com", nil)
		_, _ = client.Do(req)
		if len(tracer.FinishedSpans()) == 0 {
			t.Fatalf("finished span want at least  1, got %d", len(tracer.FinishedSpans()))
		}
	})

	t.Run("large request", func(t *testing.T) {
		tracer := mocktracer.New()
		client := NewClient(tracer)
		req, _ := http.NewRequest("POST", "https://baidu.com", strings.NewReader(strings.Repeat("f", 10000000)))
		_, _ = client.Do(req)
		if len(tracer.FinishedSpans()) == 0 {
			t.Fatalf("finished span want at least  1, got %d", len(tracer.FinishedSpans()))
		}
	})

}
