package otes

import (
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// hook is borrowed from https://github.com/olivere/elastic/tree/release-branch.v7/trace/opentracing
// under MIT license: https://github.com/olivere/elastic/blob/release-branch.v7/LICENSE

// Transport for tracing Elastic operations.
type Transport struct {
	rt     http.RoundTripper
	tracer opentracing.Tracer
}

// TransportOption signature for specifying options, e.g. WithRoundTripper.
type TransportOption func(t *Transport)

// WithRoundTripper specifies the http.RoundTripper to call
// next after this transport. If it is nil (default), the
// transport will use http.DefaultTransport.
func WithRoundTripper(rt http.RoundTripper) TransportOption {
	return func(t *Transport) {
		t.rt = rt
	}
}

// WithTracer specifies the opentracing.Tracer to call
// this transport.
func WithTracer(tracer opentracing.Tracer) TransportOption {
	return func(t *Transport) {
		t.tracer = tracer
	}
}

// NewTransport specifies a transport that will trace Elastic
// and report back via OpenTracing.
func NewTransport(opts ...TransportOption) *Transport {
	t := &Transport{}
	for _, o := range opts {
		o(t)
	}
	return t
}

// RoundTrip captures the request and starts an OpenTracing span
// for Elastic PerformRequest operation.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	span, ctx := opentracing.StartSpanFromContext(req.Context(), "es.query")
	req = req.WithContext(ctx)
	defer span.Finish()

	ext.Component.Set(span, "github.com/olivere/elastic/v7")
	ext.HTTPUrl.Set(span, req.URL.String())
	ext.HTTPMethod.Set(span, req.Method)
	ext.PeerHostname.Set(span, req.URL.Hostname())
	ext.PeerPort.Set(span, toUint16(req.URL.Port()))

	var (
		resp *http.Response
		err  error
	)
	if t.rt != nil {
		resp, err = t.rt.RoundTrip(req)
	} else {
		resp, err = http.DefaultTransport.RoundTrip(req)
	}
	if err != nil {
		ext.LogError(span, err)
	}
	if resp != nil {
		ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	}

	return resp, err
}

func toUint16(s string) uint16 {
	v, _ := strconv.ParseUint(s, 10, 16)
	return uint16(v)
}
