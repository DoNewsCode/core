package clihttp

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
)

func Example() {
	client := NewClient(opentracing.GlobalTracer())
	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	resp, _ := client.Do(req)
	resp.Body.Close()
}
