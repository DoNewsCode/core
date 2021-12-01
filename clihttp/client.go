/*
Package clihttp adds opentracing support to http client.
*/
package clihttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/DoNewsCode/core/contract"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
)

// HttpDoer modules a upstream http client.
type HttpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a http client that traces http requests.
type Client struct {
	tracer               opentracing.Tracer
	underlying           contract.HttpDoer
	requestLogThreshold  int
	responseLogThreshold int
}

// Option changes the behavior of Client.
type Option func(*Client)

// WithDoer is an option that accepts a HttpDoer as the underlying client.
func WithDoer(doer contract.HttpDoer) Option {
	return func(client *Client) {
		client.underlying = doer
	}
}

// WithRequestLogThreshold is options that sets threshold of request logging in number of bytes.
// If the payload is larger than this threshold, the log will be omit.
func WithRequestLogThreshold(num int) Option {
	return func(client *Client) {
		client.requestLogThreshold = num
	}
}

// WithResponseLogThreshold is options that sets threshold of response logging in number of bytes.
// If the response body is larger than this threshold, the log will be omit.
func WithResponseLogThreshold(num int) Option {
	return func(client *Client) {
		client.responseLogThreshold = num
	}
}

// NewClient creates a Client with tracing support.
func NewClient(tracer opentracing.Tracer, options ...Option) *Client {
	baseClient := &http.Client{Transport: &nethttp.Transport{}}
	c := &Client{
		tracer:               tracer,
		underlying:           baseClient,
		requestLogThreshold:  5000,
		responseLogThreshold: 5000,
	}
	for _, f := range options {
		f(c)
	}
	return c
}

// Do sends the request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	clientSpan, ctx := opentracing.StartSpanFromContextWithTracer(req.Context(), c.tracer, "HTTP Client")
	defer clientSpan.Finish()

	req = req.WithContext(ctx)

	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, req.RequestURI)
	ext.HTTPMethod.Set(clientSpan, req.Method)

	// Inject the client span context into the headers
	c.logRequest(req, clientSpan)

	c.tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	response, err := c.underlying.Do(req)
	if err != nil {
		ext.LogError(clientSpan,err)
		return response, err
	}

	c.logResponse(response, clientSpan)

	return response, err
}

func (c *Client) logRequest(req *http.Request, span opentracing.Span) {
	if req.Body == nil || c.requestLogThreshold <= 0 {
		return
	}
	body, err := req.GetBody()
	if err != nil {
		ext.LogError(span, errors.Wrap(err, "cannot get request body"))
		return
	}
	defer body.Close()

	byt, err := ioutil.ReadAll(io.LimitReader(body, int64(c.responseLogThreshold)))
	if err != nil {
		ext.LogError(span, errors.Wrap(err, "cannot read request body"))
		return
	}
	if span != nil {
		span.LogKV("request", string(byt))
	}

}

func (c *Client) logResponse(response *http.Response, span opentracing.Span) {
	if response.Body == nil || c.responseLogThreshold <= 0 {
		return
	}
	byt, err := ioutil.ReadAll(io.LimitReader(response.Body, int64(c.responseLogThreshold)))
	if err != nil {
		ext.LogError(span, err)
		return
	}
	if span != nil {
		span.LogKV("response", string(byt))
	}
	response.Body = readCloser{
		Closer: response.Body,
		Reader: io.MultiReader(bytes.NewReader(byt), response.Body),
	}
}

type readCloser struct {
	io.Closer
	io.Reader
}
